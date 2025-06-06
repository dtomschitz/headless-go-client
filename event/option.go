package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dtomschitz/headless-go-client/logger"
)

type ServiceOption func(context.Context, *Service) (string, error)

func WithFlushInterval(flushInterval time.Duration) ServiceOption {
	return func(ctx context.Context, service *Service) (string, error) {
		if flushInterval <= 0 {
			return "WithFlushInterval", fmt.Errorf("flush interval must be greater than 0")
		}

		service.interval = flushInterval
		return "WithFlushInterval", nil
	}
}

func WithRequestBuilder(builder RequestBuilder) ServiceOption {
	return func(ctx context.Context, s *Service) (string, error) {
		if builder == nil {
			return "WithRequestBuilder", fmt.Errorf("request builder cannot be nil")
		}
		s.requestBuilder = builder
		return "WithRequestBuilder", nil
	}
}

func WithLogger(factory logger.Factory) ServiceOption {
	return func(ctx context.Context, s *Service) (string, error) {
		if factory == nil {
			return "WitLogger", errors.New("logger is not provided")
		}

		s.logger = factory(ctx)
		return "WitLogger", nil
	}
}
