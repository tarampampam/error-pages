package static_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/http/handlers/static"
)

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	var handler = static.New([]byte{1, 2, 3})

	t.Run("get", func(t *testing.T) {
		var (
			req = httptest.NewRequest(http.MethodGet, "http://testing", http.NoBody)
			rr  = httptest.NewRecorder()
		)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Header().Get("Content-Type"), "application/octet-stream")
		assert.Equal(t, rr.Code, http.StatusOK)
		assert.Equal(t, rr.Body.Bytes(), []byte{1, 2, 3})
	})

	t.Run("head", func(t *testing.T) {
		var (
			req = httptest.NewRequest(http.MethodHead, "http://testing", http.NoBody)
			rr  = httptest.NewRecorder()
		)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, http.StatusOK)
		assert.Empty(t, rr.Header().Get("Content-Type"))
		assert.Empty(t, rr.Body.Bytes())
	})

	t.Run("method not allowed", func(t *testing.T) {
		for _, method := range []string{
			http.MethodDelete,
			http.MethodPatch,
			http.MethodPost,
			http.MethodPut,
		} {
			var (
				req = httptest.NewRequest(method, "http://testing", http.NoBody)
				rr  = httptest.NewRecorder()
			)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, rr.Header().Get("Content-Type"), "text/plain; charset=utf-8")
			assert.Equal(t, rr.Code, http.StatusMethodNotAllowed)
			assert.Equal(t, "Method Not Allowed\n", rr.Body.String())
		}
	})
}

func TestServeHTTP_Favicon(t *testing.T) {
	t.Parallel()

	var (
		handler = static.New(static.Favicon)

		req = httptest.NewRequest(http.MethodGet, "http://testing", http.NoBody)
		rr  = httptest.NewRecorder()
	)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Header().Get("Content-Type"), "image/x-icon")
	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.Bytes(), static.Favicon)
}