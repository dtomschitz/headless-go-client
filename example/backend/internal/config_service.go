package internal

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type (
	ConfigService struct {
		repository ConfigRepository
	}

	ConfigRepository interface {
		Create(ctx context.Context, config *Config) error
		GetByVersion(ctx context.Context, version string) (*Config, error)
		GetLatest(ctx context.Context) (*Config, error)
	}

	Config struct {
		Id         string                 `json:"id"`
		Version    string                 `json:"version"`
		Properties map[string]interface{} `json:"properties"`

		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
)

func NewConfigService(repository ConfigRepository) *ConfigService {
	return &ConfigService{
		repository: repository,
	}
}

func (s *ConfigService) CreateConfig(ctx context.Context, config *Config) error {
	if config.Version == "" {
		return errors.New("config version cannot be empty")
	}

	existingConfig, err := s.repository.GetByVersion(ctx, config.Version)
	if err != nil {
		return fmt.Errorf("failed to check existing config: %w", err)
	}
	if existingConfig != nil {
		return fmt.Errorf("config with version '%s' already exists", config.Version)
	}

	return s.repository.Create(ctx, config)
}

func (s *ConfigService) GetConfigByVersion(ctx context.Context, version string) (*Config, error) {
	if version == "" {
		return nil, errors.New("version cannot be empty")
	}

	config, err := s.repository.GetByVersion(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return config, nil
}

func (s *ConfigService) GetLatestConfig(ctx context.Context) (*Config, error) {
	config, err := s.repository.GetLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest config: %w", err)
	}

	return config, nil
}
