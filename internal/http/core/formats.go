package core

import (
	"bytes"
	"sort"
	"strconv"

	"github.com/valyala/fasthttp"
)

type ContentType = byte

const (
	UnknownContentType ContentType = iota // should be first
	JSONContentType
	XMLContentType
	HTMLContentType
	PlainTextContentType
)

func ClientWantFormat(ctx *fasthttp.RequestCtx) ContentType {
	// parse "Content-Type" header (e.g.: `application/json;charset=UTF-8`)
	if ct := bytes.ToLower(ctx.Request.Header.ContentType()); len(ct) > 4 { //nolint:gomnd
		return mimeTypeToContentType(ct)
	}

	// parse `X-Format` header (aka `Accept`) for the Ingress support
	// e.g.: `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8`
	if h := bytes.ToLower(bytes.TrimSpace(ctx.Request.Header.Peek(FormatHeader))); len(h) > 2 { //nolint:gomnd,nestif
		type format struct {
			mimeType []byte
			weight   float32
		}

		var formats = make([]format, 0, 8) //nolint:gomnd

		for _, b := range bytes.FieldsFunc(h, func(r rune) bool { return r == ',' }) {
			if idx := bytes.Index(b, []byte(";q=")); idx > 0 && idx < len(b) {
				f := format{b[0:idx], 0}

				if len(b) > idx+3 {
					if weight, err := strconv.ParseFloat(string(b[idx+3:]), 32); err == nil { //nolint:gomnd
						f.weight = float32(weight)
					}
				}

				formats = append(formats, f)
			} else {
				formats = append(formats, format{b, 1})
			}
		}

		switch l := len(formats); {
		case l == 0:
			return UnknownContentType

		case l == 1:
			return mimeTypeToContentType(formats[0].mimeType)
			
		default:
			sort.SliceStable(formats, func(i, j int) bool { return formats[i].weight > formats[j].weight })
			return mimeTypeToContentType(formats[0].mimeType)
		}
	}

	return UnknownContentType
}

func mimeTypeToContentType(mimeType []byte) ContentType {
	switch {
	case bytes.Contains(mimeType, []byte("application/json")), bytes.Contains(mimeType, []byte("text/json")):
		return JSONContentType

	case bytes.Contains(mimeType, []byte("application/xml")), bytes.Contains(mimeType, []byte("text/xml")):
		return XMLContentType

	case bytes.Contains(mimeType, []byte("text/html")):
		return HTMLContentType

	case bytes.Contains(mimeType, []byte("text/plain")):
		return PlainTextContentType
	}

	return UnknownContentType
}

func SetClientFormat(ctx *fasthttp.RequestCtx, t ContentType) {
	switch t {
	case JSONContentType:
		ctx.SetContentType("application/json; charset=utf-8")

	case XMLContentType:
		ctx.SetContentType("application/xml; charset=utf-8")

	case HTMLContentType:
		ctx.SetContentType("text/html; charset=utf-8")

	case PlainTextContentType:
		ctx.SetContentType("text/plain; charset=utf-8")
	}
}
