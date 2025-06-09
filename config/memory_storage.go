package config

import (
	"context"
	"sync"
)

type InMemoryStorage struct {
	mu     sync.RWMutex
	config *Config
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

func (s *InMemoryStorage) Get(ctx context.Context) (*Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return EmptyConfig, nil
	}

	return s.config, nil
}

func (s *InMemoryStorage) Save(ctx context.Context, config *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	return nil
}
