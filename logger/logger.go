package logger

import "context"

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Factory func(ctx context.Context) Logger

func New(ctx context.Context, factory Factory) Logger {
	if factory == nil {
		return &NoOpLogger{}
	}
	return factory(ctx)
}

type NoOpLogger struct{}

func (n *NoOpLogger) Debug(msg string, args ...any) {}
func (n *NoOpLogger) Info(msg string, args ...any)  {}
func (n *NoOpLogger) Warn(msg string, args ...any)  {}
func (n *NoOpLogger) Error(msg string, args ...any) {}
