package logger

import "context"

type Logger interface {
	Trace(msg ...any)
	Tracef(format string, args ...any)
	Debug(msg ...any)
	Debugf(format string, args ...any)
	Info(msg ...any)
	Infof(format string, args ...any)
	Warn(msg ...any)
	Warnf(format string, args ...any)
	Error(msg ...any)
	Errorf(format string, args ...any)
}

type Factory func(ctx context.Context) Logger

func New(ctx context.Context, factory Factory) Logger {
	if factory == nil {
		return &NoOpLogger{}
	}
	return factory(ctx)
}

type NoOpLogger struct{}

func (n *NoOpLogger) Trace(msg ...any)                  {}
func (n *NoOpLogger) Tracef(format string, args ...any) {}
func (n *NoOpLogger) Debug(msg ...any)                  {}
func (n *NoOpLogger) Debugf(format string, args ...any) {}
func (n *NoOpLogger) Info(msg ...any)                   {}
func (n *NoOpLogger) Infof(format string, args ...any)  {}
func (n *NoOpLogger) Warn(msg ...any)                   {}
func (n *NoOpLogger) Warnf(format string, args ...any)  {}
func (n *NoOpLogger) Error(msg ...any)                  {}
func (n *NoOpLogger) Errorf(format string, args ...any) {}
