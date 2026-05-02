package httpserver_test

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/httpserver"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, httpserver.New(http.NewServeMux()))
}

func TestNew_WithOptions(t *testing.T) {
	t.Parallel()

	for name, opts := range map[string][]httpserver.Option{
		"read timeout/positive":        {httpserver.WithReadTimeout(10 * time.Second)},
		"read timeout/zero":            {httpserver.WithReadTimeout(0)},
		"read timeout/negative":        {httpserver.WithReadTimeout(-1)},
		"write timeout/positive":       {httpserver.WithWriteTimeout(10 * time.Second)},
		"write timeout/zero":           {httpserver.WithWriteTimeout(0)},
		"write timeout/negative":       {httpserver.WithWriteTimeout(-1)},
		"idle timeout/positive":        {httpserver.WithIdleTimeout(10 * time.Second)},
		"idle timeout/zero":            {httpserver.WithIdleTimeout(0)},
		"idle timeout/negative":        {httpserver.WithIdleTimeout(-1)},
		"read header timeout/positive": {httpserver.WithReadHeaderTimeout(10 * time.Second)},
		"read header timeout/zero":     {httpserver.WithReadHeaderTimeout(0)},
		"read header timeout/negative": {httpserver.WithReadHeaderTimeout(-1)},
		"shutdown timeout/positive":    {httpserver.WithShutdownTimeout(10 * time.Second)},
		"shutdown timeout/zero":        {httpserver.WithShutdownTimeout(0)},
		"shutdown timeout/negative":    {httpserver.WithShutdownTimeout(-1)},
		"error log/nil":                {httpserver.WithErrorLog(nil)},
		"error log/non-nil":            {httpserver.WithErrorLog(log.New(io.Discard, "", 0))},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.NotNil(t, httpserver.New(http.NewServeMux(), opts...))
		})
	}
}

func TestServe_ContextCancel(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	addr := ln.Addr().String()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	srv := httpserver.New(handler, httpserver.WithShutdownTimeout(100*time.Millisecond))

	serveDone := make(chan error, 1)

	go func() { serveDone <- srv.Serve(ctx, ln) }()

	resp, httpErr := http.Get("http://" + addr) //nolint:gosec
	if httpErr != nil {
		t.Fatalf("HTTP request failed: %v", httpErr)
	}

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, readErr := io.ReadAll(resp.Body)
	assert.NoError(t, readErr)
	assert.Equal(t, "ok", string(body))

	cancel()

	assert.NoError(t, <-serveDone)
}

func TestServe_ErrAlreadyStarted(t *testing.T) {
	t.Parallel()

	ln1, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	defer func() { _ = ln2.Close() }() //nolint:errcheck

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // pre-cancel so Serve shuts down immediately after starting

	srv := httpserver.New(http.NewServeMux(), httpserver.WithShutdownTimeout(time.Second))

	assert.NoError(t, srv.Serve(ctx, ln1))
	assert.ErrorIs(t, srv.Serve(ctx, ln2), httpserver.ErrServerAlreadyStarted)
}

func TestServe_ListenerError(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	assert.NoError(t, ln.Close()) // close before passing to Serve to trigger a listener error

	srv := httpserver.New(http.NewServeMux())

	serveErr := srv.Serve(t.Context(), ln)
	assert.Error(t, serveErr)
	assert.NotErrorIs(t, serveErr, httpserver.ErrServerAlreadyStarted)
}
