package logreq_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"gh.tarampamp.am/error-pages/internal/http/httptest"
	"gh.tarampamp.am/error-pages/internal/http/middleware/logreq"
	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestNew(t *testing.T) {
	t.Parallel()

	var (
		buf    bytes.Buffer
		log, _ = logger.New(logger.DebugLevel, logger.JSONFormat, &buf)

		mw     = logreq.New(log, nil)
		req, _ = http.NewRequest(http.MethodPut, "http://testing/foo/bar", http.NoBody)
	)

	req.Header.Set("User-Agent", "test")
	req.Header.Set("Referer", "https://example.com")
	req.Header.Set("Content-Type", "application/json")

	httptest.HandleFastRequest(t,
		mw(func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(http.StatusOK) }),
		req,
		func(status int, body string, _ http.Header) { assert.Equal(t, http.StatusOK, status) },
	)

	var logRecord = buf.String()

	assert.Contains(t, logRecord, `"level":"info"`)
	assert.Contains(t, logRecord, `"msg":"HTTP request processed"`)
	assert.Contains(t, logRecord, `"useragent":"test"`)
	assert.Contains(t, logRecord, `"method":"PUT"`)
	assert.Contains(t, logRecord, `"url":"/foo/bar"`)
	assert.Contains(t, logRecord, `"referer":"https://example.com"`)
	assert.Contains(t, logRecord, `application/json`)
}
