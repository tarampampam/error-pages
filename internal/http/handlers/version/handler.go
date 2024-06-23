package version

import (
	"encoding/json"
	"net/http"
	"strings"
)

// New creates a handler that returns the version of the service in JSON format.
func New(ver string) http.Handler {
	var body, _ = json.Marshal(struct { //nolint:errchkjson
		Version string `json:"version"`
	}{
		Version: strings.TrimSpace(ver),
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)

		case http.MethodHead:
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
}
