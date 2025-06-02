package config

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type FileStorage struct {
	path string
	mu   sync.RWMutex
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{path: path}
}

func (fs *FileStorage) Get(ctx context.Context) (*Config, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (fs *FileStorage) Set(ctx context.Context, config *Config) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	tmpFile := fs.path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, fs.path)
}
