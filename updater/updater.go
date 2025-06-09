package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	commonCtx "github.com/dtomschitz/headless-go-client/context"
	"github.com/dtomschitz/headless-go-client/event"
	commonHttp "github.com/dtomschitz/headless-go-client/http"
	"github.com/dtomschitz/headless-go-client/logger"
)

type (
	Updater struct {
		currentVersion   string
		manifestURL      string
		initialPollDelay time.Duration
		pollInterval     time.Duration

		logger            logger.Logger
		events            event.Emitter
		updateRequester   UpdateRequester
		manifestRequester ManifestRequester

		updateAvailableChan chan *Manifest
		updateAppliedChan   chan *Manifest

		internalCtx    context.Context
		internalCancel context.CancelFunc
		wg             sync.WaitGroup
		shutdownOnce   sync.Once
	}

	UpdateEventFunc func(ctx context.Context, mainfest *Manifest)
)

const (
	ServiceName = "UpdateService"

	UpdateAvailableEvent       event.EventType = "update_available"
	NoUpdateAvailableEvent     event.EventType = "no_update_available"
	UpdateStartedEvent         event.EventType = "update_started"
	UpdateDownloadStartedEvent event.EventType = "update_download_started"
	UpdateDownloadedEvent      event.EventType = "update_downloaded"
	UpdateAppliedEvent         event.EventType = "update_applied"
)

func NewService(ctx context.Context, currentClientVersion string, opts ...Option) (*Updater, error) {
	internalCtx, internalCancel := context.WithCancel(ctx)
	internalCtx = context.WithValue(internalCtx, commonCtx.ServiceKey, ServiceName)

	if currentClientVersion == "" {
		currentClientVersion = commonCtx.GetStringValue(internalCtx, commonCtx.ClientVersionKey)
		if currentClientVersion == "" {
			internalCancel()
			return nil, fmt.Errorf("current client version cannot be empty")
		}
	}

	httpClient := commonHttp.NewClient()

	updater := &Updater{
		currentVersion:      currentClientVersion,
		updateRequester:     &DefaultUpdateRequester{Client: httpClient},
		manifestRequester:   &DefaultManifestRequester{client: httpClient},
		initialPollDelay:    1 * time.Minute,
		pollInterval:        1 * time.Hour,
		logger:              &logger.NoOpLogger{},
		events:              &event.NoopEmitter{},
		updateAvailableChan: make(chan *Manifest, 1),
		updateAppliedChan:   make(chan *Manifest, 1),
		internalCtx:         internalCtx,
		internalCancel:      internalCancel,
	}

	for _, opt := range opts {
		if err := opt(internalCtx, updater); err != nil {
			internalCancel()
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	updater.start(internalCtx)
	updater.logger.Info("started service successfully", "pollInterval", updater.pollInterval)

	return updater, nil
}

func (updater *Updater) start(ctx context.Context) {
	updater.wg.Add(1)

	go func() {
		defer updater.wg.Done()

		if updater.initialPollDelay > 0 {
			updater.logger.Info("waiting for initial poll delay before starting service", "initialPollDelay", updater.initialPollDelay)

			select {
			case <-ctx.Done():
				updater.logger.Warn("stopped service because context was cancelled")
				return
			case <-time.After(updater.initialPollDelay):
				updater.logger.Info("initial poll delay completed, starting service")
			}
		}

		ticker := time.NewTicker(updater.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				updater.logger.Warn("stopped service because context was cancelled")
				return
			case <-ticker.C:
				if err := updater.TriggerUpdateCheck(ctx); err != nil {
					updater.logger.Error("failed to trigger update check", "error", err)
					return
				}
			}
		}
	}()
}

func (updater *Updater) Name() string {
	return ServiceName
}

func (updater *Updater) Close(ctx context.Context) error {
	updater.shutdownOnce.Do(func() {
		if updater.internalCancel != nil {
			updater.internalCancel()
		}
		close(updater.updateAvailableChan)
		close(updater.updateAppliedChan)
	})

	done := make(chan struct{})
	go func() {
		updater.wg.Wait()
		close(done)
	}()

	<-done
	return nil
}

func (updater *Updater) PollEvents() []*event.Event {
	return updater.events.PollEvents()
}

func (updater *Updater) ListenForUpdateAvailable(ctx context.Context, fn UpdateEventFunc) {
	updater.eventListener(ctx, updater.updateAvailableChan, fn)
}

func (updater *Updater) ListenForUpdateApplied(ctx context.Context, fn UpdateEventFunc) {
	updater.eventListener(ctx, updater.updateAppliedChan, fn)
}

func (updater *Updater) eventListener(ctx context.Context, updateChan chan *Manifest, fn UpdateEventFunc) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case manifest := <-updateChan:
				if manifest != nil {
					fn(ctx, manifest)
				}
			}
		}
	}()
}

func (updater *Updater) TriggerUpdateCheck(ctx context.Context) error {
	manifest, isAvailable, err := updater.checkIfUpdateIsAvailable(ctx)
	if err != nil {
		updater.events.Push(event.NewEventFromError(ctx, UpdateAvailableEvent, err))
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !isAvailable {
		updater.events.Push(event.NewEvent(ctx, NoUpdateAvailableEvent))
		updater.logger.Info("no update is available")
		return nil
	}

	updater.events.Push(event.NewEvent(ctx, UpdateAvailableEvent, event.WithDataField("manifest", manifest)))
	updater.updateAvailableChan <- manifest

	return nil
}

func (updater *Updater) ApplyUpdate(ctx context.Context, manifest *Manifest) error {
	eventOpts := event.WithDataField("manifest", manifest)
	updater.events.Push(event.NewEvent(ctx, UpdateStartedEvent, eventOpts))

	if err := updater.applyUpdate(ctx, manifest); err != nil {
		err = fmt.Errorf("failed to apply update: %w", err)
		updater.events.Push(event.NewEventFromError(ctx, UpdateAppliedEvent, err, eventOpts))
		return err
	}

	updater.events.Push(event.NewEvent(ctx, UpdateAppliedEvent, eventOpts))
	updater.logger.Info("new update has been applied: %s", manifest.Version)

	return nil
}

func (updater *Updater) applyUpdate(ctx context.Context, manifest *Manifest) error {
	updater.logger.Info("going to apply update", "version", manifest.Version)
	updater.events.Push(event.NewEvent(ctx, UpdateDownloadStartedEvent))

	binary, err := updater.updateRequester.Fetch(ctx, manifest)
	if err != nil {
		return fmt.Errorf("failed to fetch update %s: %w", manifest.Version, err)
	}
	defer binary.Close()

	updater.events.Push(event.NewEvent(ctx, UpdateDownloadedEvent))
	updater.logger.Debug("update fetched successfully", "version", manifest.Version)

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
			return fmt.Errorf("updated stopped because checksum mismatch: expected %s, actual %s", manifest.SHA256, actualHash)
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
