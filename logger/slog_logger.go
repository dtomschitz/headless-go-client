package logger

import (
	"context"
	"log/slog"
	"os"

	commonCtx "github.com/dtomschitz/headless-go-client/common/context"
)

func SlogFactory(ctx context.Context) Logger {
	baseLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	var attrs []any
	if service := commonCtx.GetStringValue(ctx, commonCtx.ServiceKey); service != "" {
		attrs = append(attrs, slog.String(string(commonCtx.ServiceKey), service))
	}
	if deviceId := commonCtx.GetStringValue(ctx, commonCtx.DeviceIdKey); deviceId != "" {
		attrs = append(attrs, slog.String(string(commonCtx.DeviceIdKey), deviceId))
	}
	if clientVersion := commonCtx.GetStringValue(ctx, commonCtx.ClientVersionKey); clientVersion != "" {
		attrs = append(attrs, slog.String(string(commonCtx.ClientVersionKey), clientVersion))
	}

	return baseLogger.With(attrs...)
}
