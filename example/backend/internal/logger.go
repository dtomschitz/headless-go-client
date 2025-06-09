package internal

import (
	"context"

	"go.uber.org/zap"
)

func NewLogger(ctx context.Context) *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return logger.Sugar()
}
