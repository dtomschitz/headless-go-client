package logger

import (
	"context"
	"log/slog"
	"os"

	commonCtx "github.com/dtomschitz/headless-go-client/context"
)

func SlogFactory(ctx context.Context) Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if service := commonCtx.GetStringValue(ctx, commonCtx.ServiceKey); service != "" {
		logger = logger.With(string(commonCtx.ServiceKey), service)
	}
	if deviceId := commonCtx.GetStringValue(ctx, commonCtx.DeviceIdKey); deviceId != "" {
		logger = logger.With(string(commonCtx.DeviceIdKey), deviceId)
	}
	if clientVersion := commonCtx.GetStringValue(ctx, commonCtx.ClientVersionKey); clientVersion != "" {
		logger = logger.With(string(commonCtx.DeviceIdKey), clientVersion)
	}

	return logger
}
