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
		return &NoopLogger{}
	}
	return factory(ctx)
}

type NoopLogger struct{}

func (n *NoopLogger) Debug(msg string, args ...any) {}
func (n *NoopLogger) Info(msg string, args ...any)  {}
func (n *NoopLogger) Warn(msg string, args ...any)  {}
func (n *NoopLogger) Error(msg string, args ...any) {}
