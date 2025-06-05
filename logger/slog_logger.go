package logger

import (
	"context"
	"log/slog"
	"os"

	commonCtx "github.com/dtomschitz/headless-go-client/context"
)

func SlogFactory(ctx context.Context) Logger {
	service := commonCtx.GetStringValue(ctx, commonCtx.ServiceKey)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	return logger.With(string(commonCtx.ServiceKey), service)
}
