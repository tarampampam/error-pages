package error_page

import (
	"net/http"
	"strconv"
	"strings"
)

// getCodeFromRequest extracts the error code from the given request. It checks the URL path:
//
//   - /											-> 0, false
//   - /{code}.{ext}					-> {code}, true
//   - /{code}								-> {code}, true
//   - /{code}/{anything}			-> {code}, true
//   - /{anything}						-> 0, false
//   - /{anything}.{ext}			-> 0, false
//   - /{anything}/{anything}	-> 0, false
//
// and the headers:
//   - X-Code: {code}					-> {code}, true
//
// The code must be a number between 1 and 999 (inclusive). URL path takes priority over headers.
func getCodeFromRequest(r *http.Request) (uint16, bool) {
	// try the first URL path segment first: "/404/page" -> "404", "/404.json" -> "404"
	segment, _, _ := strings.Cut(strings.TrimLeft(r.URL.Path, "/"), "/")

	// strip extension without case-folding: numeric codes have no case
	if i := strings.LastIndexByte(segment, '.'); i >= 0 {
		segment = segment[:i]
	}

	if code, err := strconv.ParseUint(segment, 10, 16); err == nil && code > 0 && code <= 999 {
		return uint16(code), true
	}

	// fall back to the X-Code header (ingress-nginx sets this on error responses)
	if v := r.Header.Get("X-Code"); len(v) > 0 && len(v) <= 3 {
		if code, err := strconv.ParseUint(v, 10, 16); err == nil && code > 0 && code <= 999 {
			return uint16(code), true
		}
	}

	return 0, false
}
