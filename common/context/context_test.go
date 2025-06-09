package context_test

import (
	"context"
	"testing"

	commonCtx "github.com/dtomschitz/headless-go-client/common/context"
	"github.com/stretchr/testify/assert"
)

func Test_GetStringValue(t *testing.T) {
	t.Run("WithValidKey", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), commonCtx.ClientVersionKey, "1.0.0")
		value := commonCtx.GetStringValue(ctx, commonCtx.ClientVersionKey)
		assert.Equal(t, "1.0.0", value, "expected value to match the context value")
	})

	t.Run("WithInvalidKey", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), commonCtx.ClientVersionKey, "1.0.0")
		value := commonCtx.GetStringValue(ctx, "invalid_key")
		assert.Equal(t, "", value, "expected empty string for invalid key")
	})

	t.Run("WithEmptyContext", func(t *testing.T) {
		ctx := context.Background()
		value := commonCtx.GetStringValue(ctx, commonCtx.ClientVersionKey)
		assert.Equal(t, "", value, "expected empty string for empty context")
	})

	t.Run("WithNonStringValue", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), commonCtx.ClientVersionKey, 12345)
		value := commonCtx.GetStringValue(ctx, commonCtx.ClientVersionKey)
		assert.Equal(t, "", value, "expected empty string for non-string value")
	})
}
