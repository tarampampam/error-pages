package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// ErrServerAlreadyStarted is returned by [Server.Serve] when called more than once on the same instance.
// A [Server] is single-use: once started (regardless of whether it's still running or has been stopped),
// it cannot be started again. Create a new [Server] instead.
var ErrServerAlreadyStarted = errors.New("httpserver: server already started")

// Server represents an HTTP server with configurable options.
type Server struct {
	srv             *http.Server
	shutdownTimeout time.Duration
	started         atomic.Bool
}

// Option allows to configure the [Server] with functional options.
type Option func(*Server)

// WithReadTimeout sets the maximum duration for reading the entire request, including the body.
// A zero value disables the timeout. Negative values are normalized to zero.
func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		if d < 0 {
			d = 0 // no read timeout, Slowloris protection is disabled
		}

		s.srv.ReadTimeout = d
	}
}

// WithWriteTimeout sets the maximum duration before timing out writes of the response.
// A zero value disables the timeout. Negative values are normalized to zero.
//
// Warning: do not set this if your handlers may legitimately take longer than the timeout to respond -
// the connection will be forcibly closed mid-response.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		if d < 0 {
			d = 0
		}

		s.srv.WriteTimeout = d
	}
}

// WithIdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are enabled.
// Negative values are normalized to zero, in which case net/http falls back to ReadTimeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(s *Server) {
		if d < 0 {
			d = 0 // fallback to ReadTimeout (net/http's default behavior when IdleTimeout is not set)
		}

		s.srv.IdleTimeout = d
	}
}

// WithReadHeaderTimeout sets the amount of time allowed to read request headers. Default is 5 seconds.
// Negative values are normalized to zero, which disables the timeout.
//
// Warning: disabling this timeout makes the server vulnerable to Slowloris-style attacks.
// Keep a non-zero value unless you have protection upstream (e.g. a reverse proxy).
func WithReadHeaderTimeout(d time.Duration) Option {
	return func(s *Server) {
		if d < 0 {
			d = 0 // no read header timeout, vulnerable to Slowloris attacks
		}

		s.srv.ReadHeaderTimeout = d
	}
}

// WithShutdownTimeout sets the maximum amount of time to wait for in-flight requests to complete before the server
// is forcefully closed. Default is 5 seconds.
//
// Negative values are normalized to zero, which causes the server to close immediately without waiting for in-flight
// requests to complete.
func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) {
		if d < 0 {
			d = 0
		}

		s.shutdownTimeout = d
	}
}

// WithErrorLog sets the logger for errors accepting connections and unexpected behavior from handlers.
// A nil logger is ignored, leaving the previously configured (or default) logger in place.
//
// Note: [log.Logger] is used here instead of [slog.Logger] or something else due to the fact that [http.Server]
// requires a [log.Logger] for its ErrorLog field.
func WithErrorLog(logger *log.Logger) Option {
	return func(s *Server) {
		if logger == nil {
			return
		}

		s.srv.ErrorLog = logger
	}
}

// New creates a new [Server] instance with the provided options.
func New(handler http.Handler, opts ...Option) *Server {
	const (
		defaultReadHeaderTimeout = 5 * time.Second
		defaultShutdownTimeout   = 5 * time.Second
	)

	srv := &Server{
		srv: &http.Server{
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			Handler:           handler,
		},
		shutdownTimeout: defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

// Serve starts the HTTP server. It listens on the provided listener and serves incoming requests.
// To stop the server, cancel the provided context.
//
// It blocks until the context is canceled or the server is stopped by some error.
//
// The provided listener will be closed when the server stops.
//
// Serve must not be called more than once per [Server] instance. Subsequent calls return [ErrServerAlreadyStarted]
// immediately without affecting the running server (if any).
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	if !s.started.CompareAndSwap(false, true) {
		return ErrServerAlreadyStarted
	}

	errCh := make(chan error, 1)

	// closing buffered channel is not required here - GC will take care of it, but prefer it as a good practice
	go func() { defer close(errCh); errCh <- s.srv.Serve(ln) }()

	select {
	case <-ctx.Done():
		// the parent ctx is already canceled here, so we use [context.WithoutCancel] to detach from it before applying
		// WithTimeout - otherwise shutdownCtx would be canceled immediately and Shutdown would return without waiting
		// for in-flight requests. This ctx serves only as Shutdown's own drain deadline; it is not propagated to handlers
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.shutdownTimeout)
		defer cancel()

		shutdownErr := s.srv.Shutdown(shutdownCtx)
		if errors.Is(shutdownErr, context.DeadlineExceeded) {
			// if the shutdown timeout is exceeded, forcefully close the server to terminate any ongoing requests
			if closeErr := s.srv.Close(); closeErr != nil {
				shutdownErr = fmt.Errorf(
					"failed to shutdown gracefully within the timeout: %w (also failed to forcefully close the server: %w)",
					shutdownErr, closeErr,
				)
			}
		}

		serveErr := <-errCh // wait for the server to stop and capture any error from Serve

		// [http.ErrServerClosed] is the expected sentinel returned from Serve after a graceful shutdown -
		// it carries no useful information, so we drop it to keep the joined error clean
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}

		switch { // I'm not using [errors.Join] here because joined errors are separated by newlines
		case shutdownErr != nil && serveErr != nil:
			return fmt.Errorf("shutdown failed: %w (serve also returned: %w)", shutdownErr, serveErr)
		case shutdownErr != nil:
			return shutdownErr
		case serveErr != nil:
			return serveErr
		}
	case err := <-errCh:
		// ignore [http.ErrServerClosed] error, which is returned when the server is shut down gracefully
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}
