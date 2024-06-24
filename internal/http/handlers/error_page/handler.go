package error_page

import (
	"fmt"
	"net/http"

	"gh.tarampamp.am/error-pages/internal/config"
)

func New(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var code uint16

		if fromUrl, okUrl := extractCodeFromURL(r.URL.Path); okUrl {
			code = fromUrl
		} else if fromHeader, okHeaders := extractCodeFromHeaders(r.Header); okHeaders {
			code = fromHeader
		} else {
			code = cfg.DefaultCodeToRender
		}

		var httpCode int

		if cfg.RespondWithSameHTTPCode {
			httpCode = int(code)
		} else {
			httpCode = http.StatusOK
		}

		var format = detectPreferredFormatForClient(r.Header)

		switch headerName := "Content-Type"; format {
		case jsonFormat:
			w.Header().Set(headerName, "application/json; charset=utf-8")
		case xmlFormat:
			w.Header().Set(headerName, "application/xml; charset=utf-8")
		case htmlFormat:
			w.Header().Set(headerName, "text/html; charset=utf-8")
		case plainTextFormat:
			w.Header().Set(headerName, "text/plain; charset=utf-8")
		default:
			w.Header().Set(headerName, "text/html; charset=utf-8")
		}

		// https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag
		// disallow indexing of the error pages
		w.Header().Set("X-Robots-Tag", "noindex")

		if code >= 500 && code < 600 {
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
			// tell the client (search crawler) to retry the request after 120 seconds, it makes sense for the 5xx errors
			w.Header().Set("Retry-After", "120")
		}

		for _, proxyHeader := range cfg.ProxyHeaders {
			if value := r.Header.Get(proxyHeader); value != "" {
				w.Header().Set(proxyHeader, value)
			}
		}

		w.WriteHeader(httpCode)
		_, _ = w.Write([]byte(fmt.Sprintf("<html>error page for the code %d</html>", code)))
	})
}
