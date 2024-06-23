package live

import (
	"net/http"
)

// New creates a new handler that returns "OK" for GET and HEAD requests.
func New() http.Handler {
	var body = []byte("OK\n")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)

		case http.MethodHead:
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
}
