package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	commonCtx "github.com/dtomschitz/headless-go-client/context"
	"github.com/dtomschitz/headless-go-client/logger"
	"io"
	"net/http"
	"os"
	"time"
)

type (
	Updater struct {
		currentVersion string
		manifestURL    string

		logger logger.Logger

		updateRequester   UpdateRequester
		manifestRequester ManifestRequester

		initialPollDelay time.Duration
		pollInterval     time.Duration

		updateAvailableChan chan *Manifest
		updateAppliedChan   chan *Manifest
	}

	UpdateEvent struct {
		Version string
	}

	Option func(context.Context, *Updater) error
)

func WithUpdateRequester(requester UpdateRequester) Option {
	return func(ctx context.Context, updater *Updater) error {
		if requester == nil {
			return nil
		}
		updater.updateRequester = requester
		return nil
	}
}

func WithManifestRequester(requester ManifestRequester) Option {
	return func(ctx context.Context, updater *Updater) error {
		if requester == nil {
			return nil
		}
		updater.manifestRequester = requester
		return nil
	}
}

func WithPollInterval(d time.Duration) Option {
	return func(ctx context.Context, updater *Updater) error {
		if d <= 0 {
			return fmt.Errorf("poll interval must be greater than 0")
		}

		updater.pollInterval = d
		return nil
	}
}

func WithInitialPollDelay(d time.Duration) Option {
	return func(ctx context.Context, updater *Updater) error {
		if d < 0 {
			return fmt.Errorf("initial poll delay cannot be negative")
		}

		updater.initialPollDelay = d
		return nil
	}
}

func WithLogger(l logger.Logger) Option {
	return func(ctx context.Context, updater *Updater) error {
		if l == nil {
			return fmt.Errorf("logger cannot be nil")
		}

		updater.logger = l
		return nil
	}
}

func Start(ctx context.Context, currentClientVersion string, opts ...Option) (*Updater, error) {
	if currentClientVersion == "" {
		currentClientVersion = commonCtx.GetStringValue(ctx, commonCtx.ClientVersion)
		if currentClientVersion == "" {
			return nil, fmt.Errorf("current client version cannot be empty")
		}
	}

	httpClient := &http.Client{}

	updater := &Updater{
		currentVersion:      currentClientVersion,
		updateRequester:     &DefaultUpdateRequester{Client: httpClient},
		manifestRequester:   &DefaultManifestRequester{client: httpClient},
		initialPollDelay:    1 * time.Minute,
		pollInterval:        1 * time.Hour,
		logger:              logger.New(ctx, nil),
		updateAvailableChan: make(chan *Manifest),
		updateAppliedChan:   make(chan *Manifest),
	}

	for _, opt := range opts {
		if err := opt(ctx, updater); err != nil {
			return updater, err
		}
	}

	return updater, updater.start(ctx)
}

func (updater *Updater) start(ctx context.Context) error {
	if updater.updateRequester == nil {
		return fmt.Errorf("updater requester cannot be nil")
	}

	if updater.initialPollDelay > 0 {
		updater.logger.Infof("waiting for initial poll delay of %s before starting self updater", updater.initialPollDelay)
		select {
		case <-ctx.Done():
			updater.logger.Warn("self updater stopped because context was cancelled")
			return ctx.Err()
		case <-time.After(updater.initialPollDelay):
			updater.logger.Info("initial poll delay completed, starting self updater")
		}
	}

	ticker := time.NewTicker(updater.pollInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				updater.logger.Warn("self updater stopped because context was cancelled")
				return
			case <-ticker.C:
				if err := updater.TriggerUpdateCheck(ctx); err != nil {
					updater.logger.Errorf("failed to trigger update check: %v", err)
					return
				}
			}
		}
	}()

	updater.logger.Infof("self updater started successfully with poll interval of %s", updater.pollInterval)
	return nil
}

func (updater *Updater) ListenForUpdateAvailable(ctx context.Context, fn func(ctx context.Context, manifest *Manifest)) {
	updater.eventListener(ctx, updater.updateAvailableChan, func(ctx context.Context, manifest *Manifest) {
		if manifest == nil {
			return
		}
		updater.logger.Infof("new update is available: %s", manifest.Version)
		fn(ctx, manifest)
	})
}

func (updater *Updater) ListenForUpdateApplied(ctx context.Context, fn func(ctx context.Context, manifest *Manifest)) {
	updater.eventListener(ctx, updater.updateAppliedChan, func(ctx context.Context, manifest *Manifest) {
		if manifest == nil {
			return
		}
		updater.logger.Infof("new update has been applied: %s", manifest.Version)
		fn(ctx, manifest)
	})
}

func (updater *Updater) eventListener(ctx context.Context, updateChan chan *Manifest, fn func(ctx context.Context, manifest *Manifest)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case manifest := <-updateChan:
				fn(ctx, manifest)
			}
		}
	}()
}

func (updater *Updater) TriggerUpdateCheck(ctx context.Context) error {
	manifest, isAvailable, err := updater.checkIfUpdateIsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !isAvailable {
		updater.logger.Info("no update is available")
		return nil
	}

	updater.updateAvailableChan <- manifest
	return nil
}

func (updater *Updater) ApplyUpdate(ctx context.Context, manifest *Manifest) error {
	updater.logger.Infof("going to apply update with version %s", manifest.Version)

	binary, err := updater.updateRequester.Fetch(ctx, manifest)
	if err != nil {
		return fmt.Errorf("failed to fetch update %s: %w", manifest.Version, err)
	}
	defer binary.Close()

	updater.logger.Debugf("update with version %s fetched successfully", manifest.Version)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find current binary: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "update-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, hasher)

	if _, err := io.Copy(multiWriter, binary); err != nil {
		return fmt.Errorf("failed to write binary to temp file: %w", err)
	}

	if manifest.SHA256 != "" {
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if actualHash != manifest.SHA256 {
			return fmt.Errorf("updated stopped because checksum mismatch: expected %s, got %s", manifest.SHA256, actualHash)
		}
		updater.logger.Debug("going to proceed with update because checksum matches")
	}

	if err := replaceBinary(execPath, tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to replace current binary: %w", err)
	}

	updater.updateAppliedChan <- manifest

	return nil
}

func (updater *Updater) checkIfUpdateIsAvailable(ctx context.Context) (*Manifest, bool, error) {
	manifest, err := updater.manifestRequester.Fetch(ctx, updater.manifestURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	if manifest.Version == updater.currentVersion {
		return manifest, false, nil
	}

	return manifest, true, nil
}

func replaceBinary(currentPath, newBinaryPath string) error {
	backupPath := currentPath + ".bak"
	if err := os.Rename(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := os.Rename(newBinaryPath, currentPath); err != nil {
		_ = os.Rename(backupPath, currentPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	_ = os.Remove(backupPath)

	if err := os.Chmod(currentPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}
