package updater

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	commonCtx "github.com/dtomschitz/headless-go-client/common/context"
	commonHttp "github.com/dtomschitz/headless-go-client/common/http"
	"github.com/dtomschitz/headless-go-client/event"
	"github.com/dtomschitz/headless-go-client/logger"
	"github.com/dtomschitz/headless-go-client/manifest"
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
		manifestRequester manifest.ManifestRequester

		updateAvailableChan chan *manifest.Manifest
		updateAppliedChan   chan *manifest.Manifest

		internalCtx    context.Context
		internalCancel context.CancelFunc
		wg             sync.WaitGroup
		shutdownOnce   sync.Once
	}

	UpdateEventFunc func(ctx context.Context, mainfest *manifest.Manifest)
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

func NewService(ctx context.Context, manifestURL string, currentClientVersion string, opts ...Option) (*Updater, error) {
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
		manifestURL:         manifestURL,
		updateRequester:     &DefaultUpdateRequester{Client: httpClient},
		manifestRequester:   manifest.NewDefaultManifestRequester(httpClient),
		initialPollDelay:    1 * time.Minute,
		pollInterval:        1 * time.Hour,
		logger:              &logger.NoopLogger{},
		events:              &event.NoopEmitter{},
		updateAvailableChan: make(chan *manifest.Manifest, 1),
		updateAppliedChan:   make(chan *manifest.Manifest, 1),
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

		if err := updater.TriggerUpdateCheck(ctx); err != nil {
			updater.logger.Error("initial update check failed", "error", err)
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

func (updater *Updater) eventListener(ctx context.Context, updateChan chan *manifest.Manifest, fn UpdateEventFunc) {
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

func (updater *Updater) ApplyUpdate(ctx context.Context, manifest *manifest.Manifest) error {
	eventOpts := event.WithDataField("manifest", manifest)
	updater.events.Push(event.NewEvent(ctx, UpdateStartedEvent, eventOpts))

	if err := updater.applyUpdate(ctx, manifest); err != nil {
		err = fmt.Errorf("failed to apply update: %w", err)

		updater.logger.Error("failed to apply update", "error", err)
		updater.events.Push(event.NewEventFromError(ctx, UpdateAppliedEvent, err, eventOpts))
		return err
	}

	updater.events.Push(event.NewEvent(ctx, UpdateAppliedEvent, eventOpts))
	updater.logger.Info("new update has been applied: %s", manifest.Version)

	return nil
}

func (updater *Updater) applyUpdate(ctx context.Context, manifest *manifest.Manifest) error {
	updater.logger.Info("going to apply update", "version", manifest.Version)
	updater.events.Push(event.NewEvent(ctx, UpdateDownloadStartedEvent))

	binaryReader, err := updater.updateRequester.Fetch(ctx, manifest)
	if err != nil {
		return fmt.Errorf("failed to fetch update %s: %w", manifest.Version, err)
	}
	defer binaryReader.Close()

	binary, err := io.ReadAll(binaryReader)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	updater.events.Push(event.NewEvent(ctx, UpdateDownloadedEvent))
	updater.logger.Debug("update fetched successfully", "version", manifest.Version)

	if err := manifest.Verify(binary); err != nil {
		return fmt.Errorf("failed to verify update %s: %w", manifest.Version, err)
	}

	updater.logger.Debug("going to proceed with update because checksum matches")

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find current binary: %w", err)
	}

	updater.logger.Debug("resolved current binary path", "execPath", execPath)

	tmpFile, err := createTempBinaryFile(binary)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := replaceBinary(execPath, tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to replace current binary: %w", err)
	}

	updater.updateAppliedChan <- manifest

	return nil
}

func (updater *Updater) checkIfUpdateIsAvailable(ctx context.Context) (*manifest.Manifest, bool, error) {
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
		return err
	}

	_ = os.Remove(backupPath)

	if err := os.Chmod(currentPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

func createTempBinaryFile(binary []byte) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "update-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(binary); err != nil {
		return nil, fmt.Errorf("failed to write binary to temp file: %w", err)
	}
	if err := tmpFile.Chmod(0755); err != nil {
		return nil, fmt.Errorf("failed to make temp file executable: %w", err)
	}

	return tmpFile, nil
}
