package error_page

import (
	"math"
	"net/http"
	"path"
	"strconv"
	"strings"

	"gh.tarampamp.am/error-pages/v4/internal/formats"
)

// getFormatFromRequest detects the preferred response Format from the request. It checks the URL path:
//
//   - /{anything}.{ext}			-> {ext}, true
//   - /{anything}						-> "", false
//
// and the headers:
//   - Content-Type: {format}	-> {format}, true
//   - X-Format: {format}			-> {format}, true
//   - Accept: {format}				-> {format}, true
//
// Format in the URL have priority over the Format in the headers (headers are checked in the following
// order: Content-Type, X-Format, Accept).
func getFormatFromRequest(r *http.Request) (formats.Format, bool) {
	if ext := path.Ext(r.URL.Path); ext != "" {
		switch {
		case strings.EqualFold(ext, ".json"):
			return formats.JSONFormat, true
		case strings.EqualFold(ext, ".xml"):
			return formats.XMLFormat, true
		case strings.EqualFold(ext, ".html"), strings.EqualFold(ext, ".htm"):
			return formats.HTMLFormat, true
		case strings.EqualFold(ext, ".txt"):
			return formats.PlainTextFormat, true
		}
	}

	// https://developer.mozilla.org/docs/Web/HTTP/Headers/Content-Type
	//	text/html; charset=utf-8
	//	multipart/form-data; boundary=something
	//	application/json
	if ct := strings.TrimSpace(r.Header.Get("Content-Type")); ct != "" {
		// take only the first part of the content type:
		// text/html; charset=utf-8
		// ^^^^^^^^^ - will be taken
		part, _, _ := strings.Cut(ct, ";")
		if format, ok := mimeTypeToFormat(strings.TrimSpace(part)); ok {
			return format, true
		}
	}

	// https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/
	// Value of the `Accept` header sent by the client
	if xf := strings.TrimSpace(r.Header.Get("X-Format")); xf != "" {
		// ingress-nginx forwards the original client Accept header value as X-Format
		if format, ok := formatFromAcceptValue(xf); ok {
			return format, true
		}
	}

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
	//	text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8
	//	text/html
	//	image/*
	//	*/*
	//	text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
	if ah := strings.TrimSpace(r.Header.Get("Accept")); ah != "" {
		if format, ok := formatFromAcceptValue(ah); ok {
			return format, true
		}
	}

	return formats.Format(0), false
}

// formatFromAcceptValue picks the highest-weighted MIME type from an Accept-style header value
// and maps it to a Format. Wildcard entries (*/*) are ignored.
func formatFromAcceptValue(accept string) (formats.Format, bool) {
	bestMime := ""
	bestWeight := -1

	// split application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8 into parts without allocating a []string
	for accept != "" {
		segment, rest, _ := strings.Cut(accept, ",")
		accept = rest

		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		mimeType, params, _ := strings.Cut(segment, ";")

		mimeType = strings.TrimSpace(mimeType)
		if mimeType == "*/*" {
			continue // wildcard carries no format preference
		}

		if weight := parseQWeight(params); weight > bestWeight {
			bestWeight = weight
			bestMime = mimeType
		}
	}

	if bestWeight <= 0 {
		return formats.Format(0), false
	}

	return mimeTypeToFormat(bestMime)
}

// parseQWeight extracts the q quality factor from an Accept parameter string and maps it to an
// integer in [0, 10]. Returns 10 (implicit q=1.0) when no valid q parameter is present.
func parseQWeight(params string) int {
	const (
		// qScale converts float q-factors to integers: q=1.0 -> 10, q=0.9 -> 9.
		// Also serves as the default weight when no q parameter is present (implicit q=1.0).
		qScale = 10
		// qRoundFactor corrects IEEE 754 rounding in client-supplied q-values:
		// Round(q*100)/100 maps e.g. ParseFloat("0.9") -> 0.8999... to exactly 0.9.
		qRoundFactor = 100
	)

	if params == "" {
		return qScale
	}

	var qVal string

	if _, v1, ok1 := strings.Cut(params, "q="); ok1 {
		qVal = v1
	} else if _, v2, ok2 := strings.Cut(params, "Q="); ok2 {
		qVal = v2
	}

	if qVal == "" {
		return qScale
	}

	// trim any trailing parameters that follow the q-value
	if i := strings.IndexByte(qVal, ';'); i >= 0 {
		qVal = qVal[:i]
	}

	qVal = strings.TrimSpace(qVal)

	w, wErr := strconv.ParseFloat(qVal, 32)
	if wErr != nil {
		return qScale
	}

	w = math.Round(w*qRoundFactor) / qRoundFactor
	if w < 0 || w > 1 {
		return 0
	}

	return int(w * qScale)
}

// mimeTypeToFormat maps a bare MIME type (params already stripped by callers) to a Format constant.
// MIME types follow the "type/subtype" or "type/subtype+suffix" structure (RFC 6838).
func mimeTypeToFormat(mimeType string) (formats.Format, bool) {
	_, sub, ok := strings.Cut(mimeType, "/")
	if !ok {
		return formats.Format(0), false
	}

	name, suffix, _ := strings.Cut(sub, "+")

	switch {
	case strings.EqualFold(name, "json"): // application/json, text/json
		return formats.JSONFormat, true
	case strings.EqualFold(name, "xml") || strings.EqualFold(suffix, "xml"): // application/xml, application/xhtml+xml
		return formats.XMLFormat, true
	case strings.EqualFold(name, "html"): // text/html
		return formats.HTMLFormat, true
	case strings.EqualFold(name, "plain"): // text/plain
		return formats.PlainTextFormat, true
	}

	return formats.Format(0), false
}
