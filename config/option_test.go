package config

import (
	"context"
	"net/http"
	"testing"
	"time"

	require "github.com/stretchr/testify/require"
)

func TestWithEnvironmentVariables(t *testing.T) {
	// given
	ctx := context.Background()
	service := &ConfigService{}

	// when
	err := WithEnvironmentVariables()(ctx, service)

	// then
	require.NoError(t, err)
	require.True(t, service.extendWithEnvVars, "extendWithEnvVars should be true after applying option")
}

func TestWithConfigEnvPrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{
			name:    "valid prefix",
			prefix:  "CLIENT_",
			wantErr: false,
		},
		{
			name:    "empty prefix",
			prefix:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctx := context.Background()
			service := &ConfigService{envKeyPrefix: "default_"}

			// when
			err := WithConfigEnvPrefix(tt.prefix)(ctx, service)

			// then
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "env key prefix cannot be empty")
				if tt.prefix == "" {
					require.Equal(t, "default_", service.envKeyPrefix, "envKeyPrefix should retain default on error")
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.prefix, service.envKeyPrefix, "envKeyPrefix should be updated")
			}
		})
	}
}

func TestWithPollInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		wantErr  bool
	}{
		{
			name:     "valid interval",
			interval: 10 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "zero interval",
			interval: 0 * time.Second,
			wantErr:  true,
		},
		{
			name:     "negative interval",
			interval: -5 * time.Minute,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctx := context.Background()
			service := &ConfigService{pollInterval: time.Hour}

			// when
			err := WithPollInterval(tt.interval)(ctx, service)

			// then
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "poll interval must be greater than 0")
				require.Equal(t, time.Hour, service.pollInterval, "pollInterval should retain default on error")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.interval, service.pollInterval, "pollInterval should be updated")
			}
		})
	}
}

func TestWithInitialPollDelay(t *testing.T) {
	tests := []struct {
		name    string
		delay   time.Duration
		wantErr bool
	}{
		{
			name:    "valid positive delay",
			delay:   30 * time.Second,
			wantErr: false,
		},
		{
			name:    "zero delay (valid)",
			delay:   0 * time.Second,
			wantErr: false,
		},
		{
			name:    "negative delay",
			delay:   -10 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctx := context.Background()
			service := &ConfigService{initialPollDelay: 1 * time.Minute} // Default value

			// when
			err := WithInitialPollDelay(tt.delay)(ctx, service)

			// then
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "initial poll delay cannot be negative")
				require.Equal(t, 1*time.Minute, service.initialPollDelay, "initialPollDelay should retain default on error")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.delay, service.initialPollDelay, "initialPollDelay should be updated")
			}
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	tests := []struct {
		name    string
		client  *http.Client
		wantErr bool
	}{
		{
			name:    "valid HTTP client",
			client:  &http.Client{Timeout: 10 * time.Second},
			wantErr: false,
		},
		{
			name:    "nil HTTP client",
			client:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctx := context.Background()
			defaultClient := &http.Client{Timeout: 5 * time.Second}
			service := &ConfigService{client: defaultClient}

			// when
			err := WithHTTPClient(tt.client)(ctx, service)

			// then
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "http client is not provided")
				require.Equal(t, defaultClient, service.client, "client should retain default on error")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.client, service.client, "client should be updated")
			}
		})
	}
}
