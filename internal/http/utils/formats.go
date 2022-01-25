package utils

import (
	"bytes"

	"github.com/valyala/fasthttp"
)

type ContentType = byte

const (
	UnknownContentType ContentType = iota // should be first
	JSONContentType
	HTMLContentType
	PlainTextContentType
)

func ClientWantFormat(ctx *fasthttp.RequestCtx) ContentType {
	switch t := bytes.ToLower(ctx.Request.Header.Peek(fasthttp.HeaderContentType)); {
	case bytes.Contains(t, []byte("application/json")), bytes.Contains(t, []byte("text/json")):
		return JSONContentType

	case bytes.Contains(t, []byte("text/html")):
		return HTMLContentType

	case bytes.Contains(t, []byte("text/plain")):
		return PlainTextContentType
	}

	return UnknownContentType
}

func SetClientFormat(ctx *fasthttp.RequestCtx, t ContentType) {
	switch t {
	case JSONContentType:
		ctx.SetContentType("application/json; charset=utf-8")

	case HTMLContentType:
		ctx.SetContentType("text/html; charset=utf-8")

	case PlainTextContentType:
		ctx.SetContentType("text/plain; charset=utf-8")
	}
}
