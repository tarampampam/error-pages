package logreq

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// New creates a middleware for [http.ServeMux] that logs every incoming request.
//
// The skipper function should return true if the request should be skipped. It's ok to pass nil.
func New(log *zap.Logger, skipper func(*http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipper != nil && skipper(r) {
				next.ServeHTTP(w, r)

				return
			}

			var now = time.Now()

			defer func() {
				var fields = []zap.Field{
					zap.String("useragent", r.UserAgent()),
					zap.String("method", r.Method),
					zap.String("url", r.URL.String()),
					zap.String("referer", r.Referer()),
					zap.String("content type", w.Header().Get("Content-Type")),
					zap.String("remote addr", r.RemoteAddr),
					zap.String("method", r.Method),
					zap.Duration("duration", time.Since(now).Round(time.Microsecond)),
				}

				if log.Level() <= zap.DebugLevel {
					fields = append(fields,
						zap.Any("request headers", r.Header.Clone()),
						zap.Any("response headers", w.Header().Clone()),
					)
				}

				log.Info("HTTP request processed", fields...)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
