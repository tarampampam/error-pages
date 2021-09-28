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
			zap.Int("status_code", ctx.Response.StatusCode()),
			zap.Bool("connection_close", ctx.Response.ConnectionClose()),
			zap.Duration("duration", time.Since(startedAt)),
		)
	}
}
