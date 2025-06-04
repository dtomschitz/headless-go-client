package logger_test

import (
	"context"
	"testing"

	"github.com/dtomschitz/headless-go-client/logger"
	"github.com/stretchr/testify/assert"
)

type MockLogger struct{}

func (m *MockLogger) Debug(msg string, args ...any) {}
func (m *MockLogger) Info(msg string, args ...any)  {}
func (m *MockLogger) Warn(msg string, args ...any)  {}
func (m *MockLogger) Error(msg string, args ...any) {}

func TestNewLogger(t *testing.T) {
	t.Run("WithFactory", func(t *testing.T) {
		factory := func(ctx context.Context) logger.Logger {
			return &MockLogger{}
		}
		log := logger.New(context.Background(), factory)
		assert.IsType(t, &MockLogger{}, log, "expected logger to be of type MockLogger")
	})

	t.Run("WithoutFactory", func(t *testing.T) {
		log := logger.New(context.Background(), nil)
		assert.IsType(t, &logger.NoOpLogger{}, log, "expected logger to be of type NoOpLogger")
	})
}
