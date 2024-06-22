package live_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/http/handlers/live"
)

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	var (
		req = httptest.NewRequest(http.MethodGet, "http://testing", http.NoBody)
		rr  = httptest.NewRecorder()
	)

	live.New().ServeHTTP(rr, req)

	assert.Equal(t, rr.Header().Get("Content-Type"), "text/plain; charset=utf-8")
	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.String(), "OK")
}
