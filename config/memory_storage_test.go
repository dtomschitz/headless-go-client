package config

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	require "github.com/stretchr/testify/require"
)

func TestNewInMemoryStorage(t *testing.T) {
	// when
	storage := NewInMemoryStorage()

	// then
	require.NotNil(t, storage)
	require.Nil(t, storage.config, "Newly created storage should have a nil config initially")
}

func TestInMemoryStorageGetAndSave(t *testing.T) {
	// given
	storage := NewInMemoryStorage()
	ctx := context.Background()

	initialConfig, err := storage.Get(ctx)
	require.NoError(t, err)
	require.Nil(t, initialConfig, "Initial Get should return nil before any Save")

	// when
	testConfig := &Config{
		Version: "v1.0.0",
		Properties: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}
	err = storage.Save(ctx, testConfig)

	// then
	require.NoError(t, err)

	// when
	retrievedConfig, err := storage.Get(ctx)

	// then
	require.NoError(t, err)
	require.NotNil(t, retrievedConfig)
	require.Equal(t, testConfig.Version, retrievedConfig.Version, "Retrieved config version should match saved version")
	require.Equal(t, testConfig.Properties, retrievedConfig.Properties, "Retrieved config properties should match saved properties")

	require.Same(t, testConfig, retrievedConfig, "Retrieved config should be the same instance as saved")

	// when
	newConfig := &Config{
		Version: "v1.0.1",
		Properties: map[string]interface{}{
			"key3": "new_value",
		},
	}
	err = storage.Save(ctx, newConfig)

	// then
	require.NoError(t, err)

	// when
	overwrittenConfig, err := storage.Get(ctx)

	// then
	require.NoError(t, err)
	require.NotNil(t, overwrittenConfig)
	require.Equal(t, newConfig.Version, overwrittenConfig.Version, "Overwritten config version should match new version")
	require.Equal(t, newConfig.Properties, overwrittenConfig.Properties, "Overwritten config properties should match new properties")
	require.Same(t, newConfig, overwrittenConfig, "Overwritten config should be the same instance as newly saved")
}

func TestInMemoryStorageThreadSafety(t *testing.T) {
	// given
	storage := NewInMemoryStorage()
	ctx := context.Background()
	numGoroutines := 100
	numOperations := 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				config := &Config{
					Version: fmt.Sprintf("v%d.%d", gID, j),
					Properties: map[string]interface{}{
						"writer_id": gID,
						"op_num":    j,
					},
				}
				err := storage.Save(ctx, config)
				require.NoError(t, err)
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				retrievedConfig, err := storage.Get(ctx)
				require.NoError(t, err)
				require.NotNil(t, retrievedConfig)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out, possible deadlock or goroutine stuck")
	}

	finalConfig, err := storage.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, finalConfig, "Storage should contain a config after concurrent operations")
}

func TestInMemoryStorageNilConfigSave(t *testing.T) {
	// given
	storage := NewInMemoryStorage()
	ctx := context.Background()

	validConfig := &Config{Version: "v1.0", Properties: map[string]interface{}{"foo": "bar"}}
	err := storage.Save(ctx, validConfig)
	require.NoError(t, err)

	retrievedValid, err := storage.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, retrievedValid)
	require.Equal(t, "v1.0", retrievedValid.Version)

	// when
	err = storage.Save(ctx, nil)

	// then
	require.NoError(t, err)

	// when
	retrievedNil, err := storage.Get(ctx)

	// then
	require.NoError(t, err)
	require.Nil(t, retrievedNil, "Get should return nil after a nil config was saved")
}
