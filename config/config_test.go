package config

import (
	"errors"
	"fmt"
	"testing"
	"time"

	require "github.com/stretchr/testify/require"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		key        string
		want       string
		wantErr    error
	}{
		{
			name:       "Existing string key",
			properties: map[string]interface{}{"service_name": "MyService"},
			key:        "service_name",
			want:       "MyService",
			wantErr:    nil,
		},
		{
			name:       "Existing byte slice key",
			properties: map[string]interface{}{"api_token": []byte("some_token_bytes")},
			key:        "api_token",
			want:       "some_token_bytes",
			wantErr:    nil,
		},
		{
			name:       "Key not found",
			properties: map[string]interface{}{"service_name": "MyService"},
			key:        "non_existent_key",
			want:       "",
			wantErr:    ErrKeyNotFound,
		},
		{
			name:       "Wrong type - int",
			properties: map[string]interface{}{"port": 8080},
			key:        "port",
			want:       "",
			wantErr:    ErrWrongType,
		},
		{
			name:       "Wrong type - bool",
			properties: map[string]interface{}{"enabled": true},
			key:        "enabled",
			want:       "",
			wantErr:    ErrWrongType,
		},
		{
			name:       "Wrong type - float",
			properties: map[string]interface{}{"rate": 1.5},
			key:        "rate",
			want:       "",
			wantErr:    ErrWrongType,
		},
		{
			name:       "Value implementing fmt.Stringer",
			properties: map[string]interface{}{"timestamp": time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)},
			key:        "timestamp",
			want:       "2025-01-01 00:00:00 +0000 UTC",
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//given
			c := &Config{Properties: tt.properties}

			//when
			got, err := c.GetString(tt.key)

			//then
			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrKeyNotFound) || errors.Is(tt.wantErr, ErrWrongType) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		key        string
		want       int
		wantErr    error
	}{
		{
			name:       "Existing int key",
			properties: map[string]interface{}{"port": 8080},
			key:        "port",
			want:       8080,
			wantErr:    nil,
		},
		{
			name:       "Existing string convertible to int",
			properties: map[string]interface{}{"timeout": "120"},
			key:        "timeout",
			want:       120,
			wantErr:    nil,
		},
		{
			name:       "Existing float64 convertible to int",
			properties: map[string]interface{}{"rate": 99.0},
			key:        "rate",
			want:       99,
			wantErr:    nil,
		},
		{
			name:       "Existing float32 convertible to int",
			properties: map[string]interface{}{"count": float32(10)},
			key:        "count",
			want:       10,
			wantErr:    nil,
		},
		{
			name:       "Key not found",
			properties: map[string]interface{}{"port": 8080},
			key:        "non_existent_key",
			want:       0,
			wantErr:    ErrKeyNotFound,
		},
		{
			name:       "Wrong type - bool",
			properties: map[string]interface{}{"enabled": true},
			key:        "enabled",
			want:       0,
			wantErr:    ErrWrongType,
		},
		{
			name:       "Wrong type - non-numeric string",
			properties: map[string]interface{}{"version": "v1.0"},
			key:        "version",
			want:       0,
			wantErr:    fmt.Errorf("cannot convert string to int"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//given
			c := &Config{Properties: tt.properties}

			//when
			got, err := c.GetInt(tt.key)

			//then
			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrKeyNotFound) || errors.Is(tt.wantErr, ErrWrongType) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		key        string
		want       bool
		wantErr    error
	}{
		{
			name:       "Existing bool true",
			properties: map[string]interface{}{"enabled": true},
			key:        "enabled",
			want:       true,
			wantErr:    nil,
		},
		{
			name:       "Existing bool false",
			properties: map[string]interface{}{"debug_mode": false},
			key:        "debug_mode",
			want:       false,
			wantErr:    nil,
		},
		{
			name:       "String 'true'",
			properties: map[string]interface{}{"active": "true"},
			key:        "active",
			want:       true,
			wantErr:    nil,
		},
		{
			name:       "String 'false'",
			properties: map[string]interface{}{"inactive": "false"},
			key:        "inactive",
			want:       false,
			wantErr:    nil,
		},
		{
			name:       "String '1'",
			properties: map[string]interface{}{"flag_on": "1"},
			key:        "flag_on",
			want:       true,
			wantErr:    nil,
		},
		{
			name:       "String '0'",
			properties: map[string]interface{}{"flag_off": "0"},
			key:        "flag_off",
			want:       false,
			wantErr:    nil,
		},
		{
			name:       "String 'yes'",
			properties: map[string]interface{}{"consent": "yes"},
			key:        "consent",
			want:       true,
			wantErr:    nil,
		},
		{
			name:       "String 'no'",
			properties: map[string]interface{}{"opt_out": "no"},
			key:        "opt_out",
			want:       false,
			wantErr:    nil,
		},
		{
			name:       "Key not found",
			properties: map[string]interface{}{"enabled": true},
			key:        "non_existent_key",
			want:       false,
			wantErr:    ErrKeyNotFound,
		},
		{
			name:       "Wrong type - int",
			properties: map[string]interface{}{"count": 5},
			key:        "count",
			want:       false,
			wantErr:    ErrWrongType,
		},
		{
			name:       "Wrong type - non-bool string",
			properties: map[string]interface{}{"status": "pending"},
			key:        "status",
			want:       false,
			wantErr:    fmt.Errorf("cannot convert string to bool"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//given
			c := &Config{Properties: tt.properties}

			//when
			got, err := c.GetBool(tt.key)

			//then
			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrKeyNotFound) || errors.Is(tt.wantErr, ErrWrongType) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				require.False(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		key        string
		want       float64
		wantErr    error
	}{
		{
			name:       "Existing float64 key",
			properties: map[string]interface{}{"temperature": 25.5},
			key:        "temperature",
			want:       25.5,
			wantErr:    nil,
		},
		{
			name:       "Existing float32 key",
			properties: map[string]interface{}{"humidity": float32(75.3)},
			key:        "humidity",
			want:       75.3,
			wantErr:    nil,
		},
		{
			name:       "Existing int key",
			properties: map[string]interface{}{"ratio": 100},
			key:        "ratio",
			want:       100.0,
			wantErr:    nil,
		},
		{
			name:       "String convertible to float64",
			properties: map[string]interface{}{"price": "9.99"},
			key:        "price",
			want:       9.99,
			wantErr:    nil,
		},
		{
			name:       "Key not found",
			properties: map[string]interface{}{"temperature": 25.5},
			key:        "non_existent_key",
			want:       0.0,
			wantErr:    ErrKeyNotFound,
		},
		{
			name:       "Wrong type - bool",
			properties: map[string]interface{}{"active": true},
			key:        "active",
			want:       0.0,
			wantErr:    ErrWrongType,
		},
		{
			name:       "Wrong type - non-numeric string",
			properties: map[string]interface{}{"status": "error"},
			key:        "status",
			want:       0.0,
			wantErr:    fmt.Errorf("cannot convert string to float64"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//given
			c := &Config{Properties: tt.properties}

			//when
			got, err := c.GetFloat64(tt.key)

			//then
			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrKeyNotFound) || errors.Is(tt.wantErr, ErrWrongType) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.InEpsilon(t, tt.want, got, 0.0001)
			}
		})
	}
}
