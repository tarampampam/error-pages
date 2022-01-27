package breaker_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/breaker"
)

func TestNewOSSignals(t *testing.T) {
	t.Parallel()

	oss := breaker.NewOSSignals(context.Background())

	gotSignal := make(chan os.Signal, 1)

	oss.Subscribe(func(signal os.Signal) {
		gotSignal <- signal
	}, syscall.SIGUSR2)

	defer oss.Stop()

	proc, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)

	assert.NoError(t, proc.Signal(syscall.SIGUSR2)) // send the signal

	time.Sleep(time.Millisecond * 5)

	assert.Equal(t, syscall.SIGUSR2, <-gotSignal)
}

func TestNewOSSignalCtxCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	oss := breaker.NewOSSignals(ctx)

	gotSignal := make(chan os.Signal, 1)

	oss.Subscribe(func(signal os.Signal) {
		gotSignal <- signal
	}, syscall.SIGUSR2)

	defer oss.Stop()

	proc, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)

	cancel()

	assert.NoError(t, proc.Signal(syscall.SIGUSR2)) // send the signal

	assert.Empty(t, gotSignal)
}
