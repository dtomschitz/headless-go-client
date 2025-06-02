package lifecycle

import (
	"context"
	"errors"
	"sync"
)

type Closer interface {
	Close(ctx context.Context) error
}

type Manager struct {
	mu      sync.Mutex
	closers []Closer
}

func NewCloser() *Manager {
	return &Manager{}
}

func (m *Manager) Register(c Closer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closers = append([]Closer{c}, m.closers...)
}

func (m *Manager) CloseAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, c := range m.closers {
		if err := c.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
