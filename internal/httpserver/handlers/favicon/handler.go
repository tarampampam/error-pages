package favicon

import (
	_ "embed"
	"net/http"
	"strconv"
)

//go:embed favicon.ico
var faviconIco string // type string is used to tell compiler to put the content in the RO data section of the binary

// New creates a new handler that serves the favicon.ico file.
func New() http.Handler {
	body := []byte(faviconIco)
	length := strconv.Itoa(len(body))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch m := r.Method; m {
		case http.MethodGet, http.MethodHead:
			w.Header().Set("Content-Type", "image/x-icon")
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
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
