// Package breaker provides OSSignals struct for OS signals handling (with context).
package breaker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// OSSignals allows subscribing for system signals.
type OSSignals struct {
	ctx context.Context
	ch  chan os.Signal
}

// NewOSSignals creates new subscriber for system signals.
func NewOSSignals(ctx context.Context) OSSignals {
	return OSSignals{
		ctx: ctx,
		ch:  make(chan os.Signal, 1),
	}
}

// Subscribe for some system signals (call Stop for stopping).
func (oss *OSSignals) Subscribe(onSignal func(os.Signal), signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM} // default signals
	}

	signal.Notify(oss.ch, signals...)

	go func(ch <-chan os.Signal) {
		select {
		case <-oss.ctx.Done():
			break

		case sig, opened := <-ch:
			if oss.ctx.Err() != nil {
				break
			}

			if opened && sig != nil {
				onSignal(sig)
			}
		}
	}(oss.ch)
}

// Stop system signals listening.
func (oss *OSSignals) Stop() {
	signal.Stop(oss.ch)
	close(oss.ch)
}
