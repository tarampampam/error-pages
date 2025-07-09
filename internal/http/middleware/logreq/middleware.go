package logreq

import (
	"time"

	"github.com/valyala/fasthttp"

	"gh.tarampamp.am/error-pages/internal/logger"
)

// New creates a middleware that logs every incoming request.
//
// The skipper function should return true if the request should be skipped. It's ok to pass nil.
func New(
	log *logger.Logger,
	skipper func(*fasthttp.RequestCtx) bool,
) func(fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if skipper != nil && skipper(ctx) {
				next(ctx)

				return
			}

			var now = time.Now()

			defer func() {
				var fields = []logger.Attr{
					logger.Int("status code", ctx.Response.StatusCode()),
					logger.String("useragent", string(ctx.UserAgent())),
					logger.String("method", string(ctx.Method())),
					logger.String("url", string(ctx.RequestURI())),
					logger.String("referer", string(ctx.Referer())),
					logger.String("content type", string(ctx.Response.Header.ContentType())),
					logger.String("remote addr", ctx.RemoteAddr().String()),
					logger.Duration("duration", time.Since(now).Round(time.Microsecond)),
				}

				if log.Level() <= logger.DebugLevel {
					var (
						reqHeaders  = make(map[string]string)
						respHeaders = make(map[string]string)
					)

					for key, value := range ctx.Request.Header.All() {
						reqHeaders[string(key)] = string(value)
					}

					for key, value := range ctx.Response.Header.All() {
						respHeaders[string(key)] = string(value)
					}

					fields = append(fields,
						logger.Any("request headers", reqHeaders),
						logger.Any("response headers", respHeaders),
					)
				}

				log.Info("HTTP request processed", fields...)
			}()

			next(ctx)
		}
	}
}
