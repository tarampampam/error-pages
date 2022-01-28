package common

import (
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func LogRequest(h fasthttp.RequestHandler, log *zap.Logger) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var (
			startedAt = time.Now()
			ua        = string(ctx.UserAgent())
		)

		h(ctx)

		if strings.Contains(strings.ToLower(ua), "healthcheck") { // skip healthcheck requests logging
			return
		}

		log.Info("HTTP request processed",
			zap.String("useragent", ua),
			zap.String("method", string(ctx.Method())),
			zap.String("url", string(ctx.RequestURI())),
			zap.String("referer", string(ctx.Referer())),
			zap.Int("status_code", ctx.Response.StatusCode()),
			zap.String("content_type", string(ctx.Response.Header.ContentType())),
			zap.Bool("connection_close", ctx.Response.ConnectionClose()),
			zap.Duration("duration", time.Since(startedAt)),
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
