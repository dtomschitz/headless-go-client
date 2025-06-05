package lifecycle

import (
	"context"
	"errors"

	"github.com/dtomschitz/headless-go-client/logger"
)

type Option func(ctx context.Context, m *LifecycleService) error

func WithLogger(factory logger.Factory) Option {
	return func(ctx context.Context, m *LifecycleService) error {
		if factory == nil {
			return errors.New("logger is not provided")
		}
		m.logger = factory(ctx)
		return nil
	}
}
