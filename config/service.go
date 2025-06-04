package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dtomschitz/headless-go-client/logger"
)

type (
	ConfigService struct {
		url string

		initialPollDelay time.Duration
		pollInterval     time.Duration

		client *http.Client
		logger logger.Logger

		current *Config
		storage Storage
		mu      sync.RWMutex
	}
)

// ErrKeyNotFound is returned when a key does not exist.
var ErrKeyNotFound = errors.New("key not found")

// ErrWrongType is returned when the type assertion fails.
var ErrWrongType = errors.New("wrong type for key")

func NewConfigService(ctx context.Context, url string, opts ...ConfigServiceOption) (*ConfigService, error) {
	service := &ConfigService{
		url:              url,
		initialPollDelay: 1 * time.Minute,
		pollInterval:     time.Hour * 1,
		logger:           &logger.NoOpLogger{},
		client:           &http.Client{Timeout: 5 * time.Second},
		storage:          NewInMemoryStorage(),
	}

	for _, opt := range opts {
		if err := opt(ctx, service); err != nil {
			return nil, err
		}
	}

	var err error
	service.current, err = service.storage.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}

	if service.current == nil {
		if err := service.Refresh(ctx); err != nil {
			service.logger.Error("failed to load initial config from remote: %w", err)
		}
	}

	return service, service.start(ctx)
}

func (cs *ConfigService) start(ctx context.Context) error {
	if cs.initialPollDelay > 0 {
		cs.logger.Info("waiting for initial poll delay of %s before starting ConfigService", cs.initialPollDelay)
		select {
		case <-ctx.Done():
			cs.logger.Warn("ConfigService stopped because context was cancelled")
			return nil
		case <-time.After(cs.initialPollDelay):
			cs.logger.Info("initial poll delay completed, starting ConfigService")
		}
	}

	go func() {
		ticker := time.NewTicker(cs.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				cs.logger.Info("")
				return
			case <-ticker.C:
				if err := cs.Refresh(ctx); err != nil {
					cs.logger.Error("failed to refresh config: %v", err)
				}
			}
		}
	}()

	cs.logger.Info("ConfigService started successfully with poll interval of %s", cs.pollInterval)

	return nil
}

func (cs *ConfigService) Close(ctx context.Context) error {
	return nil
}

func (cs *ConfigService) Refresh(ctx context.Context) error {
	resp, err := cs.client.Get(cs.url)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var config *Config
	if err := json.Unmarshal(body, &config); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if config.Version == cs.current.Version {
		cs.logger.Debug("config version %d is up to date", config.Version)
		return nil
	}

	if err := cs.storage.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to store config: %w", err)
	}
	cs.current = config

	return nil
}

func (c *Config) GetString(key string) (string, error) {
	val, ok := c.Properties[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	switch v := val.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return "", fmt.Errorf("%w: expected string but got %T", ErrWrongType, val)
	}
}

func (c *Config) GetInt(key string) (int, error) {
	val, ok := c.Properties[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string to int: %w", err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("%w: expected int but got %T", ErrWrongType, val)
	}
}

func (c *Config) GetBool(key string) (bool, error) {
	val, ok := c.Properties[key]
	if !ok {
		return false, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	switch v := val.(type) {
	case bool:
		return v, nil
	case string:
		switch v {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no":
			return false, nil
		default:
			return false, fmt.Errorf("cannot convert string to bool: %s", v)
		}
	default:
		return false, fmt.Errorf("%w: expected bool but got %T", ErrWrongType, val)
	}
}

func (c *Config) GetFloat64(key string) (float64, error) {
	val, ok := c.Properties[key]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string to float64: %w", err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("%w: expected float64 but got %T", ErrWrongType, val)
	}
}
