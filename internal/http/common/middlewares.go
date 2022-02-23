package common

import (
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func LogRequest(h fasthttp.RequestHandler, log *zap.Logger) fasthttp.RequestHandler {
	const headersSeparator = ": "

	return func(ctx *fasthttp.RequestCtx) {
		var ua = string(ctx.UserAgent())

		if strings.Contains(strings.ToLower(ua), "healthcheck") { // skip healthcheck requests logging
			h(ctx)

			return
		}

		var reqHeaders = make([]string, 0, 24) //nolint:gomnd

		ctx.Request.Header.VisitAll(func(key, value []byte) {
			reqHeaders = append(reqHeaders, string(key)+headersSeparator+string(value))
		})

		var startedAt = time.Now()

		h(ctx)

		var respHeaders = make([]string, 0, 16) //nolint:gomnd

		ctx.Response.Header.VisitAll(func(key, value []byte) {
			respHeaders = append(respHeaders, string(key)+headersSeparator+string(value))
		})

		log.Info("HTTP request processed",
			zap.String("useragent", ua),
			zap.String("method", string(ctx.Method())),
			zap.String("url", string(ctx.RequestURI())),
			zap.String("referer", string(ctx.Referer())),
			zap.Int("status_code", ctx.Response.StatusCode()),
			zap.String("content_type", string(ctx.Response.Header.ContentType())),
			zap.Bool("connection_close", ctx.Response.ConnectionClose()),
			zap.Duration("duration", time.Since(startedAt)),
			zap.Strings("request_headers", reqHeaders),
			zap.Strings("response_headers", respHeaders),
		)
	}
}

type metrics interface {
	IncrementTotalRequests()
	ObserveRequestDuration(t time.Duration)
}

func DurationMetrics(h fasthttp.RequestHandler, m metrics) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var startedAt = time.Now()

		h(ctx)

		m.IncrementTotalRequests()
		m.ObserveRequestDuration(time.Since(startedAt))
	}
}
