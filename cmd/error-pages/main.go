package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"

	"gh.tarampamp.am/error-pages/internal/cli"
)

// set GOMAXPROCS to match Linux container CPU quota.
var _, _ = maxprocs.Set(maxprocs.Min(1), maxprocs.Logger(func(string, ...any) {}))

// main CLI application entrypoint.
func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())

		os.Exit(1)
	}
}

// run this CLI application.
func run() error {
	defer runtime.Gosched() // increase the chance of running deferred functions before exiting

	// create a context that is canceled when the user interrupts the program
	var ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return cli.NewApp(filepath.Base(os.Args[0])).Run(ctx, os.Args)
}
