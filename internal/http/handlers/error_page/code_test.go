package error_page_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"gh.tarampamp.am/error-pages/internal/http/handlers/error_page"
)

func TestURLContainsCode(t *testing.T) {
	t.Parallel()

	for giveUrl, wantOk := range map[string]bool{
		"/404":          true,
		"/404.htm":      true,
		"/404.HTM":      true,
		"/404.html":     true,
		"/404.HtmL":     true,
		"/404.css":      false,
		"/foo/404":      false,
		"/foo/404.html": false,
		"/error":        false,
		"/":             false,
		"/////":         false,
		"///404//":      false,
		"":              false,
	} {
		t.Run(giveUrl, func(t *testing.T) {
			assert.Equal(t, wantOk, error_page.URLContainsCode(giveUrl))
		})
	}
}

func TestHeadersContainCode(t *testing.T) {
	t.Parallel()

	var mkHeaders = func(key, value string) *fasthttp.RequestHeader {
		var out = new(fasthttp.RequestHeader)

		out.Set(key, value)

		return out
	}

	for name, _tt := range map[string]struct {
		giveHeaders *fasthttp.RequestHeader
		wantOk      bool
	}{
		"with code": {giveHeaders: mkHeaders("X-Code", "404"), wantOk: true},

		"empty":     {giveHeaders: nil},
		"no code":   {giveHeaders: mkHeaders("X-Code", "")},
		"wrong":     {giveHeaders: mkHeaders("X-Code", "foo")},
		"too big":   {giveHeaders: mkHeaders("X-Code", "1000")},
		"too small": {giveHeaders: mkHeaders("X-Code", "0")},
		"negative":  {giveHeaders: mkHeaders("X-Code", "-1")},
	} {
		tt := _tt

		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.wantOk, error_page.HeadersContainCode(tt.giveHeaders))
		})
	}
}
