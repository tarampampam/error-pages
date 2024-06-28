package static

import (
	_ "embed"
	"net/http"
)

//go:embed favicon.ico
var Favicon []byte

// New creates a new handler that returns the provided content for GET and HEAD requests.
func New(content []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", http.DetectContentType(content))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)

		case http.MethodHead:
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
}
