package lifecycle

import (
	"errors"

	"github.com/dtomschitz/headless-go-client/logger"
)

type Option func(*LifecycleService) error

func WithLogger(logger logger.Logger) Option {
	return func(m *LifecycleService) error {
		if logger == nil {
			return errors.New("logger is not provided")
		}
		m.logger = logger
		return nil
	}
}
