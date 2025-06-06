package config

import (
	"fmt"
	"strconv"
)

type Config struct {
	Version    string                 `json:"version"`
	Properties map[string]interface{} `json:"properties"`
}

// GetString retrieves a string value from the configuration.
// It returns an error if the key is not found or the value is not a string.
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

// GetInt retrieves an integer value from the configuration.
// It returns an error if the key is not found or the value is not an integer.
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

// GetBool retrieves a boolean value from the configuration.
// It returns an error if the key is not found or the value is not a boolean.
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

// GetFloat64 retrieves a float64 value from the configuration.
// It returns an error if the key is not found or the value is not a float64.
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
