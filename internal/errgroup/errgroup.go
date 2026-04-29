package errgroup

import (
	"context"
	"sync"
)

// Group is a collection of goroutines working on subtasks that are part of the same overall task.
type Group interface {
	// Wait blocks until all function calls from the Go method have returned, then returns the first non-nil
	// error (if any) from them.
	Wait() error

	// Go calls the given function in a new goroutine. The first call to return a non-nil error cancels the
	// group's context; that error will be returned by Wait.
	Go(f func(context.Context) error)
}

// group implementation was taken from the stdlib package golang.org/x/sync@v0.11.0, but with the following changes:
//   - SetLimit method was removed
//   - WithContext function was renamed to New
//   - New function returns [Group] (interface) instead of *group (struct)
//   - Go method accepts a function with a [context.Context] argument
type group struct {
	ctx     context.Context //nolint:containedctx
	cancel  func(error)
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

var _ Group = (*group)(nil) // ensure that group implements Group

// New returns a new group and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go returns a non-nil error or the first
// time Wait returns, whichever occurs first.
func New(ctx context.Context) (Group, context.Context) { //nolint:iface
	ctx, cancel := context.WithCancelCause(ctx)

	return &group{ctx: ctx, cancel: cancel}, ctx
}

// done decrements the WaitGroup counter.
func (g *group) done() { g.wg.Done() }

// Wait implements [Group].
func (g *group) Wait() error {
	g.wg.Wait()

	if g.cancel != nil {
		g.cancel(g.err)
	}

	return g.err
}

// Go implements [Group].
func (g *group) Go(f func(context.Context) error) {
	g.wg.Add(1)

	go func() {
		defer g.done()

		if err := f(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(g.err)
				}
			})
		}
	}()
}
