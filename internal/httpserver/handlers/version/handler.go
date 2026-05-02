package version

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// New creates a handler that returns the version of the service in JSON format.
func New(version string) http.Handler {
	body, _ := json.Marshal(struct { //nolint:errchkjson,errcheck
		Version string `json:"version"`
	}{
		Version: version,
	})
	length := strconv.Itoa(len(body))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch m := r.Method; m {
		case http.MethodGet, http.MethodHead:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
