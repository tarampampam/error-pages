package version_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/http/handlers/version"
)

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	var (
		req = httptest.NewRequest(http.MethodGet, "http://testing", http.NoBody)
		rr  = httptest.NewRecorder()
	)

	version.New("\t\n foo@bar ").ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.String(), `{"version":"foo@bar"}`)
}
