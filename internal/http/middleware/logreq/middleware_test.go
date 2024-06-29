package logreq_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/http/middleware/logreq"
	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestNew(t *testing.T) {
	t.Parallel()

	var (
		buf    bytes.Buffer
		log, _ = logger.New(logger.DebugLevel, logger.JSONFormat, &buf)

		mw  = logreq.New(log, nil)
		rr  = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPut, "/foo/bar", http.NoBody)
	)

	req.Header.Set("User-Agent", "test")
	req.Header.Set("Referer", "https://example.com")
	req.Header.Set("Content-Type", "application/json")

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	var logRecord = buf.String()

	assert.Contains(t, logRecord, `"level":"info"`)
	assert.Contains(t, logRecord, `"msg":"HTTP request processed"`)
	assert.Contains(t, logRecord, `"useragent":"test"`)
	assert.Contains(t, logRecord, `"method":"PUT"`)
	assert.Contains(t, logRecord, `"url":"/foo/bar"`)
	assert.Contains(t, logRecord, `"referer":"https://example.com"`)
	assert.Contains(t, logRecord, `application/json`)
}
