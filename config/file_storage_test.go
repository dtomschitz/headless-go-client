package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	require "github.com/stretchr/testify/require"
)

func newTempFilePath(t *testing.T) string {
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "test_config.json")
}

func TestNewFileStorage(t *testing.T) {
	// given
	testPath := "/tmp/some/path/config.json"

	// when
	storage := NewFileStorage(testPath)

	// then
	require.NotNil(t, storage)
	require.Equal(t, testPath, storage.path)
}

func TestFileStorageGetAndSet(t *testing.T) {
	// given
	filePath := newTempFilePath(t)
	storage := NewFileStorage(filePath)
	ctx := context.Background()

	initialConfig, err := storage.Get(ctx)
	require.NoError(t, err)
	require.Nil(t, initialConfig, "Initial Get should return nil before file exists")

	// when
	testConfig := &Config{
		Version: "v1.0.0",
		Properties: map[string]interface{}{
			"appName":  "TestApp",
			"logLevel": "info",
			"port":     8080,
		},
	}
	err = storage.Save(ctx, testConfig)

	// then
	require.NoError(t, err)
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var writtenConfig Config
	err = json.Unmarshal(fileContent, &writtenConfig)
	require.NoError(t, err)

	expectedPropertiesAfterUnmarshal := map[string]interface{}{
		"appName":  "TestApp",
		"logLevel": "info",
		"port":     float64(8080),
	}

	require.Equal(t, testConfig.Version, writtenConfig.Version)
	require.Equal(t, expectedPropertiesAfterUnmarshal, writtenConfig.Properties, "Properties should match after JSON unmarshal type conversion")

	// when
	retrievedConfig, err := storage.Get(ctx)

	// then
	require.NoError(t, err)
	require.NotNil(t, retrievedConfig)
	require.Equal(t, testConfig.Version, retrievedConfig.Version)
	require.Equal(t, expectedPropertiesAfterUnmarshal, retrievedConfig.Properties, "Retrieved config properties should match expected unmarshalled types")

	// Test overwriting
	// when
	newConfig := &Config{
		Version: "v1.0.1",
		Properties: map[string]interface{}{
			"appName": "NewApp",
			"env":     "production",
			"rate":    1.25,
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
	require.Equal(t, newConfig.Version, overwrittenConfig.Version)
	expectedNewPropertiesAfterUnmarshal := map[string]interface{}{
		"appName": "NewApp",
		"env":     "production",
		"rate":    1.25,
	}
	require.Equal(t, expectedNewPropertiesAfterUnmarshal, overwrittenConfig.Properties)
}

func TestFileStorageGetNonExistentFile(t *testing.T) {
	// given
	filePath := newTempFilePath(t)
	storage := NewFileStorage(filePath)
	ctx := context.Background()

	// when
	config, err := storage.Get(ctx)

	// then
	require.NoError(t, err)
	require.Nil(t, config)
}

func TestFileStorageGetInvalidJson(t *testing.T) {
	// given
	filePath := newTempFilePath(t)
	storage := NewFileStorage(filePath)
	ctx := context.Background()

	// Create a file with invalid JSON
	err := os.WriteFile(filePath, []byte(`{"version": "v1", "properties": { "key": "value"`), 0644) // Missing closing brace
	require.NoError(t, err)

	// when
	config, err := storage.Get(ctx)

	// then
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected end of JSON input")
	require.Nil(t, config)
}

func TestFileStorageSetAtomicWrite(t *testing.T) {
	// given
	filePath := newTempFilePath(t)
	storage := NewFileStorage(filePath)
	ctx := context.Background()

	initialConfig := &Config{Version: "v1.0", Properties: map[string]interface{}{"initial": "data"}}
	initialBytes, err := json.Marshal(initialConfig)
	require.NoError(t, err)
	err = os.WriteFile(filePath, initialBytes, 0644)
	require.NoError(t, err)

	newConfig := &Config{Version: "v2.0", Properties: map[string]interface{}{"new": "data"}}

	err = storage.Save(ctx, newConfig)
	require.NoError(t, err)

	tmpFilePath := filePath + ".tmp"
	_, err = os.Stat(tmpFilePath)
	require.True(t, os.IsNotExist(err), "Temporary file should not exist after successful rename")

	retrievedConfig, err := storage.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newConfig.Version, retrievedConfig.Version)
	require.Equal(t, newConfig.Properties, retrievedConfig.Properties)
}

func TestFileStorageConcurrentAccess(t *testing.T) {
	// given
	filePath := newTempFilePath(t)
	storage := NewFileStorage(filePath)
	ctx := context.Background()

	numGoroutines := 10
	numOperations := 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				configToSave := &Config{
					Version: fmt.Sprintf("v%d.%d", writerID, j),
					Properties: map[string]interface{}{
						"writer": writerID,
						"op":     j,
						"time":   time.Now().Format(time.RFC3339Nano),
					},
				}
				err := storage.Save(ctx, configToSave)
				require.NoError(t, err)
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(readerID int) {
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
	case <-time.After(10 * time.Second):
		t.Fatal("Concurrent access test timed out, potential deadlock or hang.")
	}

	finalConfig, err := storage.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, finalConfig, "Final config should not be nil after concurrent operations")
}
