package error_page

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

type preferredFormat = byte

const (
	unknownFormat   preferredFormat = iota // should be first, no format detected
	jsonFormat                             // json
	xmlFormat                              // xml
	htmlFormat                             // html
	plainTextFormat                        // plain text
)

// detectPreferredFormatForClient detects the preferred format for the client based on the headers.
// It supports the following headers: Content-Type, Accept, X-Format.
// If the headers are not set or the format is not recognized, it returns unknownFormat.
func detectPreferredFormatForClient(headers *fasthttp.RequestHeader) preferredFormat { //nolint:funlen,gocognit
	var contentType, accept string

	if contentTypeHeader := strings.TrimSpace(string(headers.Peek("Content-Type"))); contentTypeHeader != "" { //nolint:nestif,lll
		// https://developer.mozilla.org/docs/Web/HTTP/Headers/Content-Type
		//	text/html; charset=utf-8
		//	multipart/form-data; boundary=something
		//	application/json
		if parts := strings.SplitN(contentTypeHeader, ";", 2); len(parts) > 1 { //nolint:mnd
			// take only the first part of the content type:
			// text/html; charset=utf-8
			// ^^^^^^^^^ - will be taken
			contentType = strings.TrimSpace(parts[0])
		} else {
			// take the whole value
			contentType = contentTypeHeader
		}
	} else if xFormatHeader := strings.TrimSpace(string(headers.Peek("X-Format"))); xFormatHeader != "" {
		// https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/
		// Value of the `Accept` header sent by the client
		accept = xFormatHeader
	} else if acceptHeader := strings.TrimSpace(string(headers.Peek("Accept"))); acceptHeader != "" {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
		//	text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8
		//	text/html
		//	image/*
		//	*/*
		//	text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
		accept = acceptHeader
	} else {
		return unknownFormat
	}

	switch {
	case contentType != "":
		return mimeTypeToPreferredFormat(contentType)

	case accept != "":
		type piece struct {
			mimeType string
			weight   int // to avoid float32 comparison (weight 1.0 = 1_0, 0.9 = 0_9, 0.8 = 0_8, etc.)
		}

		var pieces = make([]piece, 0, strings.Count(accept, ",")+1)

		// split application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8 into parts:
		//                                                   ^^^^^^^^^ - segment #3
		//                             ^^^^^^^^^^^^^^^^^^^^^ - segment #2
		//       ^^^^^^^^^^^^^^^^^^^^^ - segment #1
		for _, segment := range strings.FieldsFunc(accept, func(r rune) bool { return r == ',' }) {
			// split segment into parts:
			//
			//	application/xhtml+xml
			//	^^^^^^^^^^^^^^^^^^^^^ - part #1
			//
			//	application/xml;q=0.9
			//	                ^^^^^ - part #2
			//	^^^^^^^^^^^^^^^ - part #1
			//
			//	*/*;q=0.8
			//	    ^^^^^ - part #2
			//	^^^ - part #1
			if parts := strings.SplitN(strings.TrimSpace(segment), ";", 2); len(parts) > 0 { //nolint:mnd,nestif
				if parts[0] == "*/*" {
					continue // skip the wildcard
				}

				var p = piece{mimeType: parts[0], weight: 1_0} //nolint:mnd // by default the weight is 10 (1.0 in float)

				if len(parts) > 1 { // we need to extract the weight
					// trim the `q=` prefix and try to parse the weight value
					if weight, err := strconv.ParseFloat(strings.TrimPrefix(strings.ToLower(parts[1]), "q="), 32); err == nil {
						if weight = math.Round(weight*100) / 100; weight <= 1 && weight >= 0 { //nolint:mnd
							p.weight = int(weight * 10) //nolint:mnd
						} else {
							p.weight = 0 // invalid weight, set it to 0
						}
					}
				}

				pieces = append(pieces, p)
			}
		}

		if len(pieces) > 0 {
			slices.SortStableFunc(pieces, func(a, b piece) int { return b.weight - a.weight })

			return mimeTypeToPreferredFormat(pieces[0].mimeType)
		}
	}

	return unknownFormat
}

// mimeTypeToPreferredFormat converts a MIME type to a preferred format, using non-string comparison.
func mimeTypeToPreferredFormat(mimeType string) preferredFormat {
	switch value := strings.ToLower(mimeType); {
	case strings.Contains(value, "/json"): // application/json text/json
		return jsonFormat
	case strings.Contains(value, "/xml"): // application/xml text/xml
		return xmlFormat
	case strings.Contains(value, "+xml"): // application/xhtml+xml
		return xmlFormat
	case strings.Contains(value, "/html"): // text/html
		return htmlFormat
	case strings.Contains(value, "/plain"): // text/plain
		return plainTextFormat
	}

	return unknownFormat
}
