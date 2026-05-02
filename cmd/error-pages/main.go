package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gh.tarampamp.am/error-pages/v4/cmd/error-pages/app"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// AFFAIR, Go runtime guarantees that os.Args[0] is always present and contains the path to the executable,
	// so we can safely use it as the application name
	return app.NewApp(filepath.Base(os.Args[0])).Run(ctx, os.Args[1:])
}
