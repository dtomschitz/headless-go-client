package updater

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dtomschitz/headless-go-client/event"
	"github.com/dtomschitz/headless-go-client/logger"
)

type Option func(context.Context, *Updater) error

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

func WithLogger(factory logger.Factory) Option {
	return func(ctx context.Context, updater *Updater) error {
		if factory == nil {
			return errors.New("logger is not provided")
		}

		updater.logger = factory(ctx)
		return nil
	}
}

func WithEventEmitter(emitter event.Emitter) Option {
	return func(ctx context.Context, updater *Updater) error {
		if emitter == nil {
			return errors.New("event emitter is not provided")
		}
		updater.events = emitter
		return nil
	}
}
