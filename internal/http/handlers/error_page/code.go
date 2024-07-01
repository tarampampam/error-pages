package error_page

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

// extractCodeFromURL extracts the error code from the given URL.
func extractCodeFromURL(url string) (uint16, bool) {
	var parts = strings.SplitN(strings.TrimLeft(url, "/"), "/", 1)

	if len(parts) == 0 {
		return 0, false
	}

	var (
		fileName = strings.ToLower(parts[0])
		ext      = filepath.Ext(fileName) // ".html", ".htm", ".%something%" or an empty string
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
func URLContainsCode(url string) (ok bool) { _, ok = extractCodeFromURL(url); return } //nolint:nlreturn

// extractCodeFromHeaders extracts the error code from the given headers.
func extractCodeFromHeaders(headers *fasthttp.RequestHeader) (uint16, bool) {
	if headers == nil {
		return 0, false
	}

	// https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/
	// HTTP status code returned by the request
	if value := headers.Peek("X-Code"); len(value) > 0 && len(value) <= 3 {
		if code, err := strconv.ParseUint(string(value), 10, 16); err == nil && code > 0 && code < 999 {
			return uint16(code), true
		}
	}

	return 0, false
}

// HeadersContainCode checks if the given headers contain an error code.
func HeadersContainCode(headers *fasthttp.RequestHeader) (ok bool) {
	_, ok = extractCodeFromHeaders(headers)

	return
}
