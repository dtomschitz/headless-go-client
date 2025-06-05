package event

import (
	"context"
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

func WithLogger(logger logger.Logger) ServiceOption {
	return func(ctx context.Context, s *Service) (string, error) {
		if logger == nil {
			return "", fmt.Errorf("logger cannot be nil")
		}

		s.logger = logger
		return "WithLogger", nil
	}
}
