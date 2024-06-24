package error_page_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

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

	for name, _tt := range map[string]struct {
		giveHeaders http.Header
		wantOk      bool
	}{
		"with code": {giveHeaders: http.Header{"X-Code": {"404"}}, wantOk: true},

		"empty":     {giveHeaders: nil},
		"no code":   {giveHeaders: http.Header{"X-Code": {""}}},
		"wrong":     {giveHeaders: http.Header{"X-Code": {"foo"}}},
		"too big":   {giveHeaders: http.Header{"X-Code": {"1000"}}},
		"too small": {giveHeaders: http.Header{"X-Code": {"0"}}},
		"negative":  {giveHeaders: http.Header{"X-Code": {"-1"}}},
	} {
		tt := _tt

		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.wantOk, error_page.HeadersContainCode(tt.giveHeaders))
		})
	}
}
