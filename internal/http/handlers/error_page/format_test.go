package error_page

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_detectPreferredFormatForClient(t *testing.T) {
	t.Parallel()

	for name, _tt := range map[string]struct {
		giveHeaders http.Header
		wantFormat  preferredFormat
	}{
		"content type json": {
			giveHeaders: http.Header{"Content-Type": {"application/jSoN"}},
			wantFormat:  jsonFormat,
		},
		"content type xml": {
			giveHeaders: http.Header{"Content-Type": {"application/xml; charset=UTF-8"}},
			wantFormat:  xmlFormat,
		},
		"content type html": {
			giveHeaders: http.Header{"Content-Type": {"text/hTmL; charset=utf-8"}},
			wantFormat:  htmlFormat,
		},
		"content type plain": {
			giveHeaders: http.Header{"Content-Type": {"text/plaIN"}},
			wantFormat:  plainTextFormat,
		},

		"accept json": {
			giveHeaders: http.Header{"Accept": {"application/jsoN,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},
		"accept xml, depends on weight": {
			giveHeaders: http.Header{"Accept": {"text/html;q=0.5,application/xhtml+xml;q=0.9,application/xml;q=1,*/*;q=0.8"}},
			wantFormat:  xmlFormat,
		},
		"accept json, depends on weight": {
			giveHeaders: http.Header{"Accept": {"application/jsoN,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},
		"accept xml": {
			giveHeaders: http.Header{"Accept": {"application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"}},
			wantFormat:  xmlFormat,
		},
		"accept html": {
			giveHeaders: http.Header{"Accept": {"text/html, application/xhtml+xml, application/xml;q=0.9, image/avif, image/webp, */*;q=0.8"}},
			wantFormat:  htmlFormat,
		},
		"accept plain": {
			giveHeaders: http.Header{"Accept": {"text/plaiN,text/html,application/xml;q=0.9,,,*/*;q=0.8"}},
			wantFormat:  plainTextFormat,
		},
		"accept json, weighted values only": {
			giveHeaders: http.Header{"Accept": {"application/jsoN;Q=0.1,text/html;q=1.1,application/xml;q=-1,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},

		"x-format json, depends on weight": {
			giveHeaders: http.Header{"X-Format": {"application/jsoN,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},
		"x-format xml": {
			giveHeaders: http.Header{"X-Format": {"application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"}},
			wantFormat:  xmlFormat,
		},

		"content type has priority over accept": {
			giveHeaders: http.Header{"Content-Type": {"text/plain"}, "Accept": {"application/xml"}},
			wantFormat:  plainTextFormat,
		},
		"accept has priority over x-format": {
			giveHeaders: http.Header{"Accept": {"application/xml"}, "X-Format": {"text/plain"}},
			wantFormat:  plainTextFormat,
		},

		"empty headers": {
			giveHeaders: nil,
		},
		"empty content type": {
			giveHeaders: http.Header{"Content-Type": {"  "}},
		},
		"wrong content type": {
			giveHeaders: http.Header{"Content-Type": {"multipart/form-data; boundary=something"}},
		},
		"wrong accept": {
			giveHeaders: http.Header{"Accept": {";q=foobar,bar/baz;;;;;application/xml"}},
		},
		"none on invalid input": {
			giveHeaders: http.Header{"Content-Type": {"foo/bar; charset=utf-8"}, "Accept": {"foo/bar; charset=utf-8"}},
		},
		"completely unknown": {
			giveHeaders: http.Header{"Content-Type": {"😀"}, "Accept": {"😄"}, "X-Format": {"😍"}},
		},
	} {
		tt := _tt

		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.wantFormat, detectPreferredFormatForClient(tt.giveHeaders))
		})
	}
}
