package middleware

import (
	"context"
	"net/http"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
)

// injectLogOnceCtxKey is a context key type for the inject log middleware to prevent duplicate injection.
type injectLogOnceCtxKey struct{}

// NewInjectLog returns a new middleware that injects the provided zap.Logger into the request context.
//
// This allows downstream handlers to retrieve the logger from the context and have access to these fields for
// logging purposes.
//
// It's safe to use this middleware multiple times in the middleware chain, as it includes a guard to prevent
// duplicate injection of the logger into the context for the same request.
func NewInjectLog(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// doubled middleware invocation guard
			if r.Context().Value(injectLogOnceCtxKey{}) != nil {
				next.ServeHTTP(w, r) // logger already injected, pass through

				return
			}

			// mark the middleware as invoked for this request to prevent duplicate injection
			ctx := context.WithValue(r.Context(), injectLogOnceCtxKey{}, struct{}{})

			// inject the logger into the context
			ctx = logger.With(ctx, log)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
