package live

import (
	"net/http"
	"strconv"
)

// New creates a new handler that always returns "OK" for GET and HEAD requests, and 405 for other methods.
// It is intended to be used as a liveness probe for Kubernetes and other orchestrators.
func New() http.Handler {
	body := []byte("OK\n")
	length := strconv.Itoa(len(body))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch m := r.Method; m {
		case http.MethodGet, http.MethodHead:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Content-Length", length)
			w.WriteHeader(http.StatusOK)

			if m == http.MethodGet {
				_, _ = w.Write(body) //nolint:errcheck
			}

		default:
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
}
