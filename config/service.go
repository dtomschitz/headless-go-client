package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"

	commonCtx "github.com/dtomschitz/headless-go-client/common/context"
	commonHttp "github.com/dtomschitz/headless-go-client/http"
	"github.com/dtomschitz/headless-go-client/logger"
	"github.com/dtomschitz/headless-go-client/manifest"
)

type (
	ConfigService struct {
		manifestURL string

		envKeyPrefix      string
		extendWithEnvVars bool
		initialPollDelay  time.Duration
		pollInterval      time.Duration

		client *http.Client
		logger logger.Logger

		manifestRequester manifest.ManifestRequester

		current *Config
		storage Storage
		mu      sync.RWMutex

		internalCtx    context.Context
		internalCancel context.CancelFunc
		wg             sync.WaitGroup
		shutdownOnce   sync.Once
	}
)

const (
	ServiceName = "ConfigService"
)

func NewService(ctx context.Context, manifestURL string, opts ...ConfigServiceOption) (*ConfigService, error) {
	internalCtx, internalCancel := context.WithCancel(ctx)
	internalCtx = context.WithValue(internalCtx, commonCtx.ServiceKey, ServiceName)

	client := commonHttp.NewClient()

	service := &ConfigService{
		manifestURL:       manifestURL,
		initialPollDelay:  1 * time.Minute,
		pollInterval:      1 * time.Hour,
		logger:            &logger.NoOpLogger{},
		client:            client,
		manifestRequester: manifest.NewDefaultManifestRequester(client),
		storage:           NewInMemoryStorage(),
		internalCtx:       internalCtx,
		internalCancel:    internalCancel,
	}

	for _, opt := range opts {
		if err := opt(internalCtx, service); err != nil {
			internalCancel()
			return nil, err
		}
	}

	var err error
	service.current, err = service.storage.Get(internalCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}

	if service.current == nil {
		if err := service.Refresh(internalCtx); err != nil {
			service.logger.Error("failed to load initial config from remote: %w", err)
		}
	}

	service.start(internalCtx)
	service.logger.Info("started service successfully", "pollInterval", service.pollInterval)

	return service, nil
}

func (cs *ConfigService) start(ctx context.Context) {
	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()

		if cs.initialPollDelay > 0 {
			cs.logger.Info("waiting for initial poll delay before starting service", "initialPollDelay", cs.initialPollDelay)
			select {
			case <-ctx.Done():
				cs.logger.Warn("stopped service because context was cancelled")
				return
			case <-time.After(cs.initialPollDelay):
				cs.logger.Info("initial poll delay completed, starting service")
			}
		}

		ticker := time.NewTicker(cs.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				cs.logger.Warn("stopped service because context was cancelled")
				return
			case <-ticker.C:
				if err := cs.Refresh(ctx); err != nil {
					cs.logger.Error("failed to refresh config: %v", err)
				}
			}
		}
	}()
}

func (cs *ConfigService) Name() string {
	return "ConfigService"
}

func (cs *ConfigService) Close(ctx context.Context) error {
	cs.shutdownOnce.Do(func() {
		if cs.internalCancel != nil {
			cs.internalCancel()
		}
	})

	done := make(chan struct{})
	go func() {
		cs.wg.Wait()
		close(done)
	}()

	<-done
	return nil
}

func (cs *ConfigService) Current() *Config {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return deepCopyConfig(cs.current)
}

func (cs *ConfigService) Refresh(ctx context.Context) error {
	return cs.refresh(ctx)
}

func (cs *ConfigService) refresh(ctx context.Context) error {
	manifest, err := cs.manifestRequester.Fetch(ctx, cs.manifestURL)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	config, err := cs.fetchFromRemote(ctx, manifest.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}

	if err := manifest.Verify(config); err != nil {
		return fmt.Errorf("failed to verify config: %w", err)
	}

	var properties Properties
	if err := json.Unmarshal(config, &properties); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	newConfig := &Config{
		Version:    manifest.Version,
		Hash:       manifest.Hash,
		Properties: properties,
	}

	if cs.extendWithEnvVars {
		cs.logger.Debug("extending config with environment variables")
		newConfig = cs.extendWithEnvironmentVariables(newConfig)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.current != nil && newConfig.Version == cs.current.Version {
		cs.logger.Info("config version is up to date", "version", newConfig.Version)
		return nil
	}

	if isConfigContentEqual(newConfig, cs.current) {
		cs.logger.Info("config properties have not changed")
		return nil
	}

	if err := cs.storage.Save(ctx, newConfig); err != nil {
		return fmt.Errorf("failed to store config: %w", err)
	}
	cs.current = newConfig

	return nil
}

func (cs *ConfigService) fetchFromRemote(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := cs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	config, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return config, nil
}

func (cs *ConfigService) extendWithEnvironmentVariables(baseConfig *Config) *Config {
	if baseConfig == nil {
		baseConfig = &Config{Properties: make(map[string]interface{})}
	} else {
		baseConfig = deepCopyConfig(baseConfig)
	}

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, cs.envKeyPrefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				cs.logger.Debug("invalid environment variable", "env", env)
				continue
			}

			envKey := parts[0]
			envValue := parts[1]

			configKey := strings.TrimPrefix(envKey, cs.envKeyPrefix)
			configKey = strings.ToLower(configKey)

			if _, ok := baseConfig.Properties[configKey]; ok {
				cs.logger.Debug("environment variable already applied", "key", configKey, "value", envKey)
				continue
			}

			baseConfig.Properties[configKey] = envValue
			cs.logger.Debug("applied environment variable override", "key", configKey, "value", envKey)
		}
	}

	return baseConfig
}

func isConfigContentEqual(a, b *Config) bool {
	return cmp.Equal(a.Properties, b.Properties)
}

func deepCopyConfig(config *Config) *Config {
	if config == nil {
		return nil
	}

	copied := &Config{
		Version:    config.Version,
		Properties: make(map[string]interface{}, len(config.Properties)),
	}
	for k, v := range config.Properties {
		copied.Properties[k] = v
	}

	return copied
}
