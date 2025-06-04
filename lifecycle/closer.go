package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dtomschitz/headless-go-client/logger"
)

type Closer interface {
	Name() string
	Close(ctx context.Context) error
}

type (
	LifecycleService struct {
		logger  logger.Logger
		mu      sync.Mutex
		closers []Closer
	}
)

func NewLifecycleService(ctx context.Context, opts ...Option) (*LifecycleService, error) {
	service := &LifecycleService{
		mu:      sync.Mutex{},
		logger:  &logger.NoOpLogger{},
		closers: make([]Closer, 0),
	}
	for _, opt := range opts {
		if err := opt(service); err != nil {
			return nil, err
		}
	}

	return service, nil
}

func (m *LifecycleService) Register(c Closer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closers = append([]Closer{c}, m.closers...)

	m.logger.Info("successfully registed closer for closer %s", c.Name())
}

func (m *LifecycleService) CloseAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, c := range m.closers {
		if err := c.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to close %s: %w", c.Name(), err))
		}
	}

	err := errors.Join(errs...)
	if err != nil {
		m.logger.Error("failed to close all closers: %v", err)
		return err
	}

	m.logger.Info("successfully closed all closers")
	return nil
}
