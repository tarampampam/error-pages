package static_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/http/handlers/static"
	"gh.tarampamp.am/error-pages/internal/http/httptest"
)

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	var (
		handler = static.New([]byte{1, 2, 3})
		url     = "http://testing"
		body    = http.NoBody
	)

	t.Run("get", func(t *testing.T) {
		httptest.HandleFast(t, handler, http.MethodGet, url, body, func(status int, body string, headers http.Header) {
			assert.Equal(t, http.StatusOK, status)
			assert.Equal(t, "application/octet-stream", headers.Get("Content-Type"))
			assert.Equal(t, []byte{1, 2, 3}, []byte(body))
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

func TestServeHTTP_Favicon(t *testing.T) {
	t.Parallel()

	httptest.HandleFast(t,
		static.New(static.Favicon),
		http.MethodGet,
		"http://testing",
		http.NoBody,
		func(status int, body string, headers http.Header) {
			assert.Equal(t, http.StatusOK, status)
			assert.Equal(t, "image/x-icon", headers.Get("Content-Type"))
			assert.Equal(t, static.Favicon, []byte(body))
		},
	)
}
