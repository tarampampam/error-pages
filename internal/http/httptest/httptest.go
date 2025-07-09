// Package httptest provides utilities for (fast-)HTTP testing.
package httptest

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

// HandleFastRequest serves http request using provided fasthttp handler and HTTP request.
func HandleFastRequest(
	t *testing.T,
	handler fasthttp.RequestHandler,
	req *http.Request,
	check func(status int, body string, _ http.Header),
) {
	t.Helper()

	// create in-memory listener
	var ln = fasthttputil.NewInmemoryListener()

	defer func() { require.NoError(t, ln.Close()) }()

	// start fasthttp server
	go func() { require.NoError(t, fasthttp.Serve(ln, handler)) }()

	// send http request
	resp, respErr := (&http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) { return ln.Dial() },
		},
	}).Do(req)
	require.NoError(t, respErr)

	// close response body after the test
	defer func() { assert.NoError(t, resp.Body.Close()) }()

	// read response body
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// check the response
	check(resp.StatusCode, string(respBody), resp.Header)
}

// HandleFast serves http request using provided fasthttp handler.
func HandleFast(
	t *testing.T,
	handler fasthttp.RequestHandler,
	method string,
	url string,
	body io.Reader,
	check func(status int, body string, _ http.Header),
) {
	t.Helper()

	// create http request
	req, reqErr := http.NewRequest(method, url, body)
	require.NoError(t, reqErr)

	// serve http request
	HandleFastRequest(t, handler, req, check)
}
