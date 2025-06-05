package config

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dtomschitz/headless-go-client/logger"
)

type ConfigServiceOption func(context.Context, *ConfigService) error

func WithPollInterval(d time.Duration) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if d <= 0 {
			return errors.New("poll interval must be greater than 0")
		}

		service.pollInterval = d
		return nil
	}
}

func WithInitialPollDelay(d time.Duration) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if d < 0 {
			return errors.New("initial poll delay cannot be negative")
		}

		service.initialPollDelay = d
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

func WithLogger(factory logger.Factory) ConfigServiceOption {
	return func(ctx context.Context, service *ConfigService) error {
		if factory == nil {
			return errors.New("logger is not provided")
		}

		service.logger = factory(ctx)
		return nil
	}
}
