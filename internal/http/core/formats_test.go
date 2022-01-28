package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/http/core"
	"github.com/valyala/fasthttp"
)

func TestClientWantFormat(t *testing.T) {
	for name, tt := range map[string]struct {
		giveContentTypeHeader string
		giveFormatHeader      string
		giveReqCtx            func() *fasthttp.RequestCtx
		wantFormat            core.ContentType
	}{
		"priority": {
			giveFormatHeader:      "application/xml",
			giveContentTypeHeader: "text/plain",
			wantFormat:            core.PlainTextContentType,
		},
		"format respects weight": {
			giveFormatHeader: "text/html;q=0.5,application/xhtml+xml;q=0.9,application/xml;q=1,*/*;q=0.8",
			wantFormat:       core.XMLContentType,
		},
		"wrong format value": {
			giveFormatHeader: ";q=foobar,bar/baz;;;;;application/xml",
			wantFormat:       core.UnknownContentType,
		},

		"content type - application/json": {
			giveContentTypeHeader: "application/jsoN; charset=utf-8", wantFormat: core.JSONContentType,
		},
		"content type - text/json": {
			giveContentTypeHeader: "text/Json; charset=utf-8", wantFormat: core.JSONContentType,
		},
		"format - json": {
			giveFormatHeader: "application/jsoN,*/*;q=0.8", wantFormat: core.JSONContentType,
		},

		"content type - application/xml": {
			giveContentTypeHeader: "application/xmL; charset=utf-8", wantFormat: core.XMLContentType,
		},
		"content type - text/xml": {
			giveContentTypeHeader: "text/Xml; charset=utf-8", wantFormat: core.XMLContentType,
		},
		"format - xml": {
			giveFormatHeader: "text/Xml", wantFormat: core.XMLContentType,
		},

		"content type - text/html": {
			giveContentTypeHeader: "text/htMl; charset=utf-8", wantFormat: core.HTMLContentType,
		},
		"format - html": {
			giveFormatHeader: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			wantFormat:       core.HTMLContentType,
		},

		"content type - text/plain": {
			giveContentTypeHeader: "text/plaiN; charset=utf-8", wantFormat: core.PlainTextContentType,
		},
		"format - plain": {
			giveFormatHeader: "text/plaiN,text/html,application/xml;q=0.9,,,*/*;q=0.8", wantFormat: core.PlainTextContentType,
		},

		"unknown on empty": {
			wantFormat: core.UnknownContentType,
		},
		"unknown on foo/bar": {
			giveContentTypeHeader: "foo/bar; charset=utf-8",
			giveFormatHeader:      "foo/bar; charset=utf-8",
			wantFormat:            core.UnknownContentType,
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := &fasthttp.RequestHeader{}
			h.Set(fasthttp.HeaderContentType, tt.giveContentTypeHeader)
			h.Set(core.FormatHeader, tt.giveFormatHeader)

			ctx := &fasthttp.RequestCtx{
				Request: fasthttp.Request{
					Header: *h, //nolint:govet
				},
			}

			assert.Equal(t, tt.wantFormat, core.ClientWantFormat(ctx))
		})
	}
}

func TestSetClientFormat(t *testing.T) {
	for name, tt := range map[string]struct {
		giveContentType core.ContentType
		wantHeaderValue string
	}{
		"plain on unknown": {giveContentType: core.UnknownContentType, wantHeaderValue: "text/plain; charset=utf-8"},
		"json":             {giveContentType: core.JSONContentType, wantHeaderValue: "application/json; charset=utf-8"},
		"xml":              {giveContentType: core.XMLContentType, wantHeaderValue: "application/xml; charset=utf-8"},
		"html":             {giveContentType: core.HTMLContentType, wantHeaderValue: "text/html; charset=utf-8"},
		"plain":            {giveContentType: core.PlainTextContentType, wantHeaderValue: "text/plain; charset=utf-8"},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{
				Response: fasthttp.Response{
					Header: fasthttp.ResponseHeader{},
				},
			}

			assert.Empty(t, "", ctx.Response.Header.Peek(fasthttp.HeaderContentType))

			core.SetClientFormat(ctx, tt.giveContentType)

			assert.Equal(t, tt.wantHeaderValue, string(ctx.Response.Header.Peek(fasthttp.HeaderContentType)))
		})
	}
}
