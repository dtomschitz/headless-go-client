package config

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dtomschitz/headless-go-client/event"
	"github.com/dtomschitz/headless-go-client/logger"
	"github.com/dtomschitz/headless-go-client/manifest"
)

type ConfigServiceOption func(context.Context, *ConfigService) error

func WithEnvironmentVariables() ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		service.extendWithEnvVars = true
		return nil
	}
}

func WithConfigEnvPrefix(prefix string) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if prefix == "" {
			return errors.New("env key prefix cannot be empty")
		}

		service.envKeyPrefix = prefix
		return nil
	}
}

func WithPollInterval(pollInterval time.Duration) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if pollInterval <= 0 {
			return errors.New("poll interval must be greater than 0")
		}

		service.pollInterval = pollInterval
		return nil
	}
}

func WithInitialPollDelay(initialPollDelay time.Duration) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if initialPollDelay < 0 {
			return errors.New("initial poll delay cannot be negative")
		}

		service.initialPollDelay = initialPollDelay
		return nil
	}
}

func WithHTTPClient(client *http.Client) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if client == nil {
			return errors.New("http client is not provided")
		}

		service.client = client
		return nil
	}
}

func WithManifestRequester(requester manifest.ManifestRequester) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if requester == nil {
			return errors.New("manifest requester is not provided")
		}
		service.manifestRequester = requester
		return nil
	}
}

func WithStorage(storage ConfigStorage) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if storage == nil {
			return errors.New("storage impl is not provided")
		}

		service.storage = storage
		return nil
	}
}

func WithEventEmitter(emitter event.Emitter) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if emitter == nil {
			return errors.New("event emitter is not provided")
		}
		service.events = emitter
		return nil
	}
}

func WithLogger(factory logger.Factory) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if factory == nil {
			return errors.New("logger is not provided")
		}

		service.logger = factory(ctx)
		return nil
	}
}
