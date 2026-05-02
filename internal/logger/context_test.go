package logger_test

import (
	"bytes"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestWith(t *testing.T) {
	t.Parallel()

	t.Run("stores logger in context", func(t *testing.T) {
		t.Parallel()

		log := logger.NewNop()
		ctx := logger.With(t.Context(), log)

		assert.Equal(t, log, logger.FromContext(ctx))
		assert.Equal(t, true, logger.IsSet(ctx))
	})
}

func TestFromContext(t *testing.T) {
	t.Parallel()

	t.Run("returns no-op logger when not present", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		log := logger.FromContext(ctx)

		assert.NotNil(t, log)
		assert.Equal(t, false, logger.IsSet(ctx))

		// must not panic when logging
		log.Info("test message")
	})
}

func TestWithFields(t *testing.T) {
	t.Parallel()

	t.Run("adds fields to logger in context", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log, err := logger.New(logger.DebugLevel, logger.JSONFormat, logger.WithWriter(&buf))
		assert.NoError(t, err)

		ctx := logger.With(t.Context(), log)
		assert.Equal(t, true, logger.IsSet(ctx))

		ctxWithFields := logger.WithFields(ctx, logger.String("key", "value"))
		logger.FromContext(ctxWithFields).Info("test message")

		assert.Contains(t, buf.String(), `"key":"value"`)
		assert.Contains(t, buf.String(), `"msg":"test message"`)
	})

	t.Run("returns original context if no logger present", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		newCtx := logger.WithFields(ctx, logger.String("key", "value"))

		assert.Equal(t, false, logger.IsSet(ctx))
		assert.Equal(t, ctx, newCtx) // same context - no logger, so WithFields is a no-op
	})

	t.Run("returns original context if no fields provided", func(t *testing.T) {
		t.Parallel()

		log := logger.NewNop()
		ctx := logger.With(t.Context(), log)

		assert.Equal(t, true, logger.IsSet(ctx))
		assert.Equal(t, ctx, logger.WithFields(ctx)) // no-op on empty attrs
	})

	t.Run("accumulates fields across calls", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log, err := logger.New(logger.DebugLevel, logger.JSONFormat, logger.WithWriter(&buf))
		assert.NoError(t, err)

		ctx := logger.With(t.Context(), log)
		ctx = logger.WithFields(ctx, logger.String("a", "1"))
		ctx = logger.WithFields(ctx, logger.String("b", "2"))

		logger.FromContext(ctx).Info("msg")

		assert.Contains(t, buf.String(), `"a":"1"`)
		assert.Contains(t, buf.String(), `"b":"2"`)
	})
}
