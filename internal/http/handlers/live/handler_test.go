package live_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/http/handlers/live"
	"gh.tarampamp.am/error-pages/internal/http/httptest"
)

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	var (
		handler = live.New()
		url     = "http://testing"
		body    = http.NoBody
	)

	t.Run("get", func(t *testing.T) {
		httptest.HandleFast(t, handler, http.MethodGet, url, body, func(status int, body string, headers http.Header) {
			assert.Equal(t, http.StatusOK, status)
			assert.Equal(t, "text/plain; charset=utf-8", headers.Get("Content-Type"))
			assert.Equal(t, "OK\n", body)
		})
	})

	t.Run("head", func(t *testing.T) {
		httptest.HandleFast(t, handler, http.MethodHead, url, body, func(status int, body string, headers http.Header) {
			assert.Equal(t, http.StatusOK, status)
			assert.Empty(t, headers.Get("Content-Type"))
			assert.Empty(t, body)
		})
	})

	t.Run("method not allowed", func(t *testing.T) {
		for _, method := range []string{
			http.MethodDelete,
			http.MethodPatch,
			http.MethodPost,
			http.MethodPut,
		} {
			httptest.HandleFast(t, handler, method, url, body, func(status int, body string, headers http.Header) {
				assert.Equal(t, http.StatusMethodNotAllowed, status)
				assert.Equal(t, "text/plain; charset=utf-8", headers.Get("Content-Type"))
				assert.Equal(t, "Method Not Allowed\n", body)
			})
		}
	})
}
