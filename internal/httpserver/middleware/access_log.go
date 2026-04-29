package middleware

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"slices"
	"sync/atomic"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
)

// accessLogOnceCtxKey is a context key type for the inject log middleware to prevent duplicate injection.
type accessLogOnceCtxKey struct{}

// NewAccessLog returns a middleware that logs HTTP requests at the specified logger.Level (http errors and server
// errors are automatically logged at higher levels). The log entry will include the request URI, method, user agent,
// host, remote address and other relevant information.
//
// It depends on the logger being injected into the request context by the injectLog middleware, so it should be used
// AFTER the NewInjectLog middleware in the middleware chain.
//
// To capture the response status code and size, this middleware wraps the [http.ResponseWriter] with a custom
// implementation, that implements the most common interfaces ([http.Flusher], [http.Hijacker], [http.Pusher]) to ensure
// compatibility with a wide range of handlers and middleware that may rely on these interfaces.
//
// The skipper function can be provided to conditionally skip logging for certain requests. If the skipper function
// returns true for a request, that request will not be logged, but the next handler in the chain will still be called.
//
// This middleware includes a guard to prevent duplicate logging for the same request, in case it is
// accidentally added multiple times in the middleware chain.
func NewAccessLog(
	lvl logger.Level, // zapcore.InfoLevel by default, because it's zero-value of the zapcore.Level
	skipper func(*http.Request) bool, // optional, may be nil
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// skip the middleware if the skipper function is provided and returns true for the request
			if skipper != nil && skipper(r) {
				next.ServeHTTP(w, r)

				return
			}

			// doubled middleware invocation guard - if the middleware has already been invoked for this request, skip it
			// to avoid duplicate logging
			if r.Context().Value(accessLogOnceCtxKey{}) != nil {
				next.ServeHTTP(w, r) // already invoked, pass through

				return
			}

			r = r.WithContext(context.WithValue(r.Context(), accessLogOnceCtxKey{}, struct{}{}))

			now := time.Now()
			rw := &accessLogResponseWriter{orig: w}

			defer func() {
				message, level := "Request successfully processed", lvl

				status := int(rw.Status.Load())
				if status == 0 {
					status = http.StatusOK
				}

				switch {
				case status >= http.StatusInternalServerError: // 500
					message, level = "Server error", logger.ErrorLevel
				case status >= http.StatusBadRequest: // 400
					message, level = "Client error", logger.WarnLevel
				case status >= http.StatusMultipleChoices: // 300
					message = "Redirection"
				case status >= http.StatusContinue && status < http.StatusOK: // 1xx
					message = "Informational"
				}

				attrs := []logger.Attr{
					logger.Duration("duration", time.Since(now).Round(time.Microsecond)),
					logger.Int("status_code", status),
					logger.Int64("response_size", rw.Size.Load()),
					logger.String("url", r.URL.String()),
					logger.String("method", r.Method),
					logger.String("useragent", r.UserAgent()),
					logger.String("referer", r.Referer()),
					logger.String("content_type", rw.Header().Get("Content-Type")),
					logger.String("host", r.Host),
					logger.String("remote_addr", r.RemoteAddr),
				}

				log := logger.FromContext(r.Context())

				// include request/response headers to attrs if the logger level is debug or lower
				if log.Level() <= logger.DebugLevel {
					attrs = append(attrs,
						logger.Strings("request_headers", headersToStrings(r.Header)...),
						logger.Strings("response_headers", headersToStrings(rw.Header())...),
					)
				}

				log.Log(level, message, attrs...)
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

// headersToStrings converts [http.Header] to a sorted slice of strings in the format "Key: Value" for logging purposes.
func headersToStrings(headers http.Header) []string {
	list := make([]string, 0, len(headers))

	for key, values := range headers {
		for _, value := range values {
			list = append(list, key+": "+value)
		}
	}

	slices.Sort(list)

	return list
}

// accessLogResponseWriter is a wrapper around [http.ResponseWriter] that captures the status code and response
// size for access logging purposes.
type accessLogResponseWriter struct {
	orig http.ResponseWriter

	// atomic fields to safely capture status and size even if the handler writes the response from multiple goroutines
	// (e.g., by using http.Flusher to flush partial responses asynchronously); this may be excessive for most use
	// cases, but it ensures correctness in all cases without relying on the handler's implementation details
	Status atomic.Int32
	Size   atomic.Int64
}

var ( // ensure accessLogResponseWriter implements the most common interfaces
	_ http.ResponseWriter                       = (*accessLogResponseWriter)(nil)
	_ interface{ Unwrap() http.ResponseWriter } = (*accessLogResponseWriter)(nil) // see http.ResponseController
	_ io.ReaderFrom                             = (*accessLogResponseWriter)(nil)
	_ http.Flusher                              = (*accessLogResponseWriter)(nil)
	_ http.Hijacker                             = (*accessLogResponseWriter)(nil)
	_ http.Pusher                               = (*accessLogResponseWriter)(nil)
)

// Header delegates to the underlying ResponseWriter's Header method.
func (rw *accessLogResponseWriter) Header() http.Header { return rw.orig.Header() }

// Write captures the size of the response body being written and ensures that if the status code has not been set
// yet, it defaults to 200 OK.
func (rw *accessLogResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.orig.Write(b)

	rw.Size.Add(int64(n))
	rw.Status.CompareAndSwap(0, http.StatusOK) // if status code has not been set yet, set it to 200 OK by default

	return n, err
}

// WriteHeader captures the status code being written and delegates to the underlying ResponseWriter's
// WriteHeader method.
func (rw *accessLogResponseWriter) WriteHeader(statusCode int) {
	rw.Status.CompareAndSwap(0, int32(statusCode)) //nolint:gosec // HTTP status codes are always within int32 range

	// we don't protect against multiple calls to WriteHeader here, as it's the responsibility of the handler to
	// call it correctly (and the default http.Server will log a warning if WriteHeader is called multiple times)
	rw.orig.WriteHeader(statusCode)
}

// Unwrap returns the original [http.ResponseWriter] that is being wrapped by accessLogResponseWriter.
func (rw *accessLogResponseWriter) Unwrap() http.ResponseWriter { return rw.orig }

// ReadFrom implements the [io.ReaderFrom] interface to efficiently capture the response size when handlers
// use [io.Copy] to write the response body.
func (rw *accessLogResponseWriter) ReadFrom(src io.Reader) (n int64, err error) {
	rw.Status.CompareAndSwap(0, http.StatusOK)

	if w, ok := rw.orig.(io.ReaderFrom); ok {
		n, err = w.ReadFrom(src)
	} else {
		n, err = io.Copy(rw.orig, src)
	}

	rw.Size.Add(n)

	return n, err
}

// Flush delegates to the underlying Flusher if present.
func (rw *accessLogResponseWriter) Flush() {
	if flusher, ok := rw.orig.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack delegates to the underlying Hijacker if present; otherwise return a wrapped error indicating the
// operation is not supported.
func (rw *accessLogResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.orig.(http.Hijacker); ok {
		return hj.Hijack()
	}

	return nil, nil, errors.New("access log middleware: underlying ResponseWriter does not implement http.Hijacker")
}

// Push delegates to the underlying Pusher if present; otherwise return an error.
//
// Note: HTTP/2 Server Push was removed from Chrome 106+ (2022) and Firefox 132+ (2024), making it effectively
// obsolete in browser contexts. [http.Pusher] is retained here solely for compatibility with handlers that may
// type-assert for it. See: https://developer.chrome.com/blog/removing-push
func (rw *accessLogResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.orig.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}

	return errors.New("access log middleware: underlying ResponseWriter does not implement http.Pusher")
}
