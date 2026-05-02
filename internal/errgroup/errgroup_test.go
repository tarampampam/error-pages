package errgroup_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/errgroup"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestGroup_Go_NothingToDo(t *testing.T) {
	t.Parallel()

	var (
		ctx, cancel = context.WithCancel(context.Background())
		eg, egCtx   = errgroup.New(ctx)
		called      bool
		ctxMatched  bool
	)

	defer cancel()

	eg.Go(func(ctx context.Context) error {
		ctxMatched = ctx == egCtx
		called = true

		return nil
	})

	assert.NoError(t, eg.Wait())
	assert.True(t, called)
	assert.True(t, ctxMatched)
}

func TestGroup_Go_CancelOnFirstError(t *testing.T) {
	t.Parallel()

	var (
		ctx, cancel = context.WithCancel(context.Background())
		eg, _       = errgroup.New(ctx)
		counter     atomic.Uint32
	)

	defer cancel()

	eg.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			counter.Add(1) // should not be reached due to context cancellation
		}

		return errors.New("long") // ignored; the first error wins
	})

	eg.Go(func(ctx context.Context) error {
		<-time.After(time.Millisecond)
		counter.Add(1)

		return errors.New("short")
	})

	assert.ErrorEqual(t, eg.Wait(), "short")
	assert.Equal(t, uint32(1), counter.Load())
}
