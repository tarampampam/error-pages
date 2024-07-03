package http_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/config"
	appHttp "gh.tarampamp.am/error-pages/internal/http"
	"gh.tarampamp.am/error-pages/internal/logger"
)

// TestRouting in fact is a test for the whole server, because it tests all the routes and their handlers.
func TestRouting(t *testing.T) {
	var (
		srv = appHttp.NewServer(logger.NewNop(), 1025*5)
		cfg = config.New()
	)

	assert.NoError(t, cfg.Templates.Add("unit-test", `<!DOCTYPE html>
<html lang="en">
	<h1>Error {{ code }}: {{ message }}</h1>{{ if description }}
	<h2>{{ description }}</h2>{{ end }}{{ if show_details }}

	<pre>
		Host: {{ host }}
		Original URI: {{ original_uri }}
		Forwarded For: {{ forwarded_for }}
		Namespace: {{ namespace }}
		Ingress Name: {{ ingress_name }}
		Service Name: {{ service_name }}
		Service Port: {{ service_port }}
		Request ID: {{ request_id }}
		Timestamp: {{ nowUnix }}
	</pre>{{ end }}
</html>`))

	cfg.TemplateName = "unit-test"

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
		t.Run("success", func(t *testing.T) {
			t.Run("index, default (plain text by default)", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/")

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "404: Not Found")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("index, default (json format)", func(t *testing.T) {
				var status, body, headers = sendRequest(t,
					http.MethodGet, baseUrl+"/", map[string]string{"Accept": "application/json"},
				)

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), `"code": 404`)
				assert.Contains(t, headers.Get("Content-Type"), "application/json")
			})

			t.Run("index, default (xml format)", func(t *testing.T) {
				var status, body, headers = sendRequest(t,
					http.MethodGet, baseUrl+"/", map[string]string{"Accept": "application/xml"},
				)

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), `<code>404</code>`)
				assert.Contains(t, headers.Get("Content-Type"), "application/xml")
			})

			t.Run("index, default (html format)", func(t *testing.T) {
				var status, body, headers = sendRequest(t,
					http.MethodGet, baseUrl+"/", map[string]string{"Content-Type": "text/html"},
				)

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), `<h1>Error 404: Not Found</h1>`)
				assert.Contains(t, headers.Get("Content-Type"), "text/html")
			})

			t.Run("index, code in HTTP header", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/", map[string]string{"X-Code": "404"})

				assert.Equal(t, http.StatusOK, status) // because of [cfg.RespondWithSameHTTPCode] is false by default
				assert.Contains(t, string(body), "404: Not Found")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("code in URL, .html", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/500.html")

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "500: Internal Server Error")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("code in URL, .htm", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/409.htm")

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "409: Conflict")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("code in URL, without extension", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/405")

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "405: Method Not Allowed")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("code in the URL have higher priority than in the headers", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/405", map[string]string{"X-Code": "404"})

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "405: Method Not Allowed")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("invalid code in HTTP header (with a string)", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/", map[string]string{"X-Code": "foobar"})

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "404: Not Found")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("invalid code in HTTP header (too small)", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/", map[string]string{"X-Code": "0"})

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "404: Not Found")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("invalid code in HTTP header (too big)", func(t *testing.T) {
				var status, body, headers = sendRequest(t, http.MethodGet, baseUrl+"/", map[string]string{"X-Code": "1000"})

				assert.Equal(t, http.StatusOK, status)
				assert.Contains(t, string(body), "404: Not Found")
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
			})

			t.Run("other HTTP methods", func(t *testing.T) {
				for _, method := range []string{http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut} {
					var status, body, headers = sendRequest(t, method, baseUrl+"/404.html")

					assert.Equal(t, http.StatusOK, status)
					assert.Contains(t, string(body), "404: Not Found")
					assert.Contains(t, headers.Get("Content-Type"), "text/plain")
				}
			})
		})

		t.Run("failure", func(t *testing.T) {
			var assertIsNotErrorPage = func(t *testing.T, body []byte) {
				t.Helper()

				assert.NotContains(t, string(body), "error page") // FIXME
			}

			t.Run("invalid code in URL (too small)", func(t *testing.T) {
				var status, body, _ = sendRequest(t, http.MethodGet, baseUrl+"/0.html")

				assert.Equal(t, http.StatusNotFound, status)
				assertIsNotErrorPage(t, body)
			})

			t.Run("invalid code in URL (too big)", func(t *testing.T) {
				var status, body, _ = sendRequest(t, http.MethodGet, baseUrl+"/1000.html")

				assert.Equal(t, http.StatusNotFound, status)
				assertIsNotErrorPage(t, body)
			})

			t.Run("invalid code in URL (with a string suffix)", func(t *testing.T) {
				var status, body, _ = sendRequest(t, http.MethodGet, baseUrl+"/404foobar.html")

				assert.Equal(t, http.StatusNotFound, status)
				assertIsNotErrorPage(t, body)
			})

			t.Run("invalid code in URL (with a string prefix)", func(t *testing.T) {
				var status, body, _ = sendRequest(t, http.MethodGet, baseUrl+"/foobar404.html")

				assert.Equal(t, http.StatusNotFound, status)
				assertIsNotErrorPage(t, body)
			})

			t.Run("invalid code in URL (with a string)", func(t *testing.T) {
				var status, body, _ = sendRequest(t, http.MethodGet, baseUrl+"/foobar.html")

				assert.Equal(t, http.StatusNotFound, status)
				assertIsNotErrorPage(t, body)
			})
		})
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
				assert.Contains(t, headers.Get("Content-Type"), "text/plain")
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

	return fmt.Sprintf("http://%s", hostPort), func() { assert.NoError(t, srv.Stop(350*time.Millisecond)) }
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
