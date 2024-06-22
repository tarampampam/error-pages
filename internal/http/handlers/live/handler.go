package live

import (
	"net/http"
)

// New creates a new handler that always returns "OK" with status code 200.
func New() http.Handler {
	var body = []byte("OK")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
}
