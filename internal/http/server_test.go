package http_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"gh.tarampamp.am/error-pages/internal/config"
	appHttp "gh.tarampamp.am/error-pages/internal/http"
)

func TestRouting(t *testing.T) {
	var (
		srv = appHttp.NewServer(context.Background(), zap.NewNop())
		cfg = config.New()
	)

	require.NoError(t, srv.Register(&cfg))

	var baseUrl, stopServer = startServer(t, &srv)

	defer stopServer()

	t.Run("health", func(t *testing.T) {
		var routes = []string{"/health/live", "/health", "/healthz", "/live"}

		t.Run("success (get)", func(t *testing.T) {
			for _, route := range routes {
				status, body, headers := sendRequest(t, http.MethodGet, baseUrl+route)

				assert.Equal(t, http.StatusOK, status)
				assert.NotEmpty(t, body)
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			}
		})

		t.Run("success (head)", func(t *testing.T) {
			for _, route := range routes {
				status, body, headers := sendRequest(t, http.MethodHead, baseUrl+route)

				assert.Equal(t, http.StatusOK, status)
				assert.Empty(t, body)
				assert.Empty(t, headers.Get("Content-Type"))
			}
		})

		t.Run("method not allowed", func(t *testing.T) {
			for _, route := range routes {
				var url = baseUrl + route

				for _, method := range []string{http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut} {
					status, body, headers := sendRequest(t, method, url)

					assert.Equal(t, http.StatusMethodNotAllowed, status)
					assert.NotEmpty(t, body)
					assert.Contains(t, headers.Get("Content-Type"), "text/plain")
				}
			}
		})
	})

	t.Run("version", func(t *testing.T) {
		var url = baseUrl + "/version"

		t.Run("success (get)", func(t *testing.T) {
			status, body, headers := sendRequest(t, http.MethodGet, url)

			assert.Equal(t, http.StatusOK, status)
			assert.NotEmpty(t, body)
			assert.Contains(t, headers.Get("Content-Type"), "application/json")
		})

		t.Run("success (head)", func(t *testing.T) {
			status, body, headers := sendRequest(t, http.MethodHead, url)

			assert.Equal(t, http.StatusOK, status)
			assert.Empty(t, body)
			assert.Empty(t, headers.Get("Content-Type"))
		})

		t.Run("method not allowed", func(t *testing.T) {
			for _, method := range []string{http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut} {
				status, body, headers := sendRequest(t, method, url)

				assert.Equal(t, http.StatusMethodNotAllowed, status)
				assert.NotEmpty(t, body)
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			}
		})
	})

	t.Run("error page", func(t *testing.T) {
		t.Skip("not implemented")
	})

	t.Run("errors handling", func(t *testing.T) {
		var missingRoutes = []string{"/not-found", "/not-found/", "/not-found.html"}

		t.Run("not found (get)", func(t *testing.T) {
			for _, path := range missingRoutes {
				status, body, headers := sendRequest(t, http.MethodGet, baseUrl+path)

				assert.Equal(t, http.StatusNotFound, status)
				assert.NotEmpty(t, body)
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			}
		})

		t.Run("not found (head)", func(t *testing.T) {
			for _, path := range missingRoutes {
				status, body, headers := sendRequest(t, http.MethodHead, baseUrl+path)

				assert.Equal(t, http.StatusNotFound, status)
				assert.Empty(t, body)
				assert.Empty(t, headers.Get("Content-Type"))
			}
		})

		t.Run("methods not allowed", func(t *testing.T) {
			for _, path := range missingRoutes {
				for _, method := range []string{http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut} {
					status, body, headers := sendRequest(t, method, baseUrl+path)

					assert.Equal(t, http.StatusMethodNotAllowed, status)
					assert.NotEmpty(t, body)
					assert.Contains(t, headers.Get("Content-Type"), "text/plain")
				}
			}
		})
	})
}

// sendRequest is a helper function to send an HTTP request and return its status code, body, and headers.
func sendRequest(t *testing.T, method, url string, headers ...map[string]string) (
	status int,
	body []byte,
	_ http.Header,
) {
	t.Helper()

	req, reqErr := http.NewRequest(method, url, nil)

	require.NoError(t, reqErr)

	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	body, _ = io.ReadAll(resp.Body)

	require.NoError(t, resp.Body.Close())

	return resp.StatusCode, body, resp.Header
}

// startServer is a helper function to start an HTTP server and return its base URL and a stop function.
func startServer(t *testing.T, srv *appHttp.Server) (_ string, stop func()) {
	t.Helper()

	var (
		port     = getFreeTcpPort(t)
		hostPort = fmt.Sprintf("%s:%d", "127.0.0.1", port)
	)

	go func() {
		if err := srv.Start("127.0.0.1", port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			assert.NoError(t, err)
		}
	}()

	// wait until the server starts
	for {
		if conn, err := net.DialTimeout("tcp", hostPort, time.Second); err == nil {
			require.NoError(t, conn.Close())

			break
		}

		<-time.After(5 * time.Millisecond)
	}

	return fmt.Sprintf("http://%s", hostPort), func() { assert.NoError(t, srv.Stop(10*time.Millisecond)) }
}

// getFreeTcpPort is a helper function to get a free TCP port number.
func getFreeTcpPort(t *testing.T) uint16 {
	t.Helper()

	l, lErr := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, lErr)

	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())

	// make sure port is closed
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			break
		}

		require.NoError(t, conn.Close())
		<-time.After(5 * time.Millisecond)
	}

	return uint16(port)
}
