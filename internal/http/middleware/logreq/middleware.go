package logreq

import (
	"net/http"
	"time"

	"gh.tarampamp.am/error-pages/internal/logger"
)

// New creates a middleware for [http.ServeMux] that logs every incoming request.
//
// The skipper function should return true if the request should be skipped. It's ok to pass nil.
func New(log *logger.Logger, skipper func(*http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipper != nil && skipper(r) {
				next.ServeHTTP(w, r)

				return
			}

			var now = time.Now()

			defer func() {
				var fields = []logger.Attr{
					logger.String("useragent", r.UserAgent()),
					logger.String("method", r.Method),
					logger.String("url", r.URL.String()),
					logger.String("referer", r.Referer()),
					logger.String("content type", w.Header().Get("Content-Type")),
					logger.String("remote addr", r.RemoteAddr),
					logger.String("method", r.Method),
					logger.Duration("duration", time.Since(now).Round(time.Microsecond)),
				}

				if log.Level() <= logger.DebugLevel {
					fields = append(fields,
						logger.Any("request headers", r.Header.Clone()),
						logger.Any("response headers", w.Header().Clone()),
					)
				}

				log.Info("HTTP request processed", fields...)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
