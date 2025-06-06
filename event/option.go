package event

import (
	"context"
	"errors"
	"fmt"

	"github.com/dtomschitz/headless-go-client/logger"
)

type ServiceOption func(context.Context, *Service) (string, error)

func WithRequestBuilder(builder RequestBuilder) ServiceOption {
	return func(ctx context.Context, s *Service) (string, error) {
		if builder == nil {
			return "", fmt.Errorf("request builder cannot be nil")
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
