package utils

import (
	"bytes"

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
	var (
		ct = bytes.ToLower(ctx.Request.Header.Peek(fasthttp.HeaderContentType))
		f  = bytes.ToLower(ctx.Request.Header.Peek(FormatHeader)) // for the Ingress support
	)

	switch {
	case bytes.Contains(f, []byte("json")),
		bytes.Contains(ct, []byte("application/json")),
		bytes.Contains(ct, []byte("text/json")):
		return JSONContentType

	case bytes.Contains(f, []byte("xml")),
		bytes.Contains(ct, []byte("application/xml")),
		bytes.Contains(ct, []byte("text/xml")):
		return XMLContentType

	case bytes.Contains(f, []byte("html")),
		bytes.Contains(ct, []byte("text/html")):
		return HTMLContentType

	case bytes.Contains(f, []byte("plain")),
		bytes.Contains(ct, []byte("text/plain")):
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
