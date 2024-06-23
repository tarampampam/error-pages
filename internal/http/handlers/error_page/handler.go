package error_page

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"gh.tarampamp.am/error-pages/internal/config"
)

func New(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var code uint16

		if fromUrl, okUrl := ExtractCodeFromURL(r.URL.Path); okUrl {
			code = fromUrl
		} else if fromHeader, okHeaders := ExtractCodeFromHeaders(r.Header); okHeaders {
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

		w.Header().Set("Content-Type", "text/html; charset=utf-8") // TODO: should depends on requested type
		w.WriteHeader(httpCode)
		_, _ = w.Write([]byte(fmt.Sprintf("<html>error page for the code %d</html>", code)))
	})
}

// ExtractCodeFromURL extracts the error code from the given URL.
func ExtractCodeFromURL(url string) (uint16, bool) {
	var parts = strings.SplitN(strings.TrimLeft(url, "/"), "/", 1)

	if len(parts) == 0 {
		return 0, false
	}

	var (
		fileName = parts[0]
		ext      = strings.ToLower(filepath.Ext(fileName)) // ".html", ".htm", ".%something%" or an empty string
	)

	if ext != "" && ext != ".html" && ext != ".htm" {
		return 0, false
	} else if ext != "" {
		fileName = strings.TrimSuffix(fileName, ext)
	}

	if code, err := strconv.ParseUint(fileName, 10, 16); err == nil && code > 0 && code < 999 {
		return uint16(code), true
	}

	return 0, false
}

// URLContainsCode checks if the given URL contains an error code.
func URLContainsCode(url string) (ok bool) { _, ok = ExtractCodeFromURL(url); return } //nolint:nlreturn

// ExtractCodeFromHeaders extracts the error code from the given headers.
func ExtractCodeFromHeaders(headers http.Header) (uint16, bool) {
	if value := headers.Get("X-Code"); len(value) > 0 && len(value) <= 3 {
		if code, err := strconv.ParseUint(value, 10, 16); err == nil && code > 0 && code < 999 {
			return uint16(code), true
		}
	}

	return 0, false
}

// HeadersContainCode checks if the given headers contain an error code.
func HeadersContainCode(headers http.Header) (ok bool) {
	_, ok = ExtractCodeFromHeaders(headers)

	return
}
