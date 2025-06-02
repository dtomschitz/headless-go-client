package config

import (
	"context"
)

type (
	Storage interface {
		Get(ctx context.Context) (*Config, error)
		Save(ctx context.Context, config *Config) error
	}
)
