package error_page

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func Test_detectPreferredFormatForClient(t *testing.T) {
	t.Parallel()

	for name, _tt := range map[string]struct {
		giveHeaders map[string][]string
		wantFormat  preferredFormat
	}{
		"content type json": {
			giveHeaders: map[string][]string{"Content-Type": {"application/jSoN"}},
			wantFormat:  jsonFormat,
		},
		"content type xml": {
			giveHeaders: map[string][]string{"Content-Type": {"application/xml; charset=UTF-8"}},
			wantFormat:  xmlFormat,
		},
		"content type html": {
			giveHeaders: map[string][]string{"Content-Type": {"text/hTmL; charset=utf-8"}},
			wantFormat:  htmlFormat,
		},
		"content type plain": {
			giveHeaders: map[string][]string{"Content-Type": {"text/plaIN"}},
			wantFormat:  plainTextFormat,
		},

		"accept json": {
			giveHeaders: map[string][]string{"Accept": {"application/jsoN,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},
		"accept xml, depends on weight": {
			giveHeaders: map[string][]string{"Accept": {"text/html;q=0.5,application/xhtml+xml;q=0.9,application/xml;q=1,*/*;q=0.8"}},
			wantFormat:  xmlFormat,
		},
		"accept json, depends on weight": {
			giveHeaders: map[string][]string{"Accept": {"application/jsoN,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},
		"accept xml": {
			giveHeaders: map[string][]string{"Accept": {"application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"}},
			wantFormat:  xmlFormat,
		},
		"accept html": {
			giveHeaders: map[string][]string{"Accept": {"text/html, application/xhtml+xml, application/xml;q=0.9, image/avif, image/webp, */*;q=0.8"}},
			wantFormat:  htmlFormat,
		},
		"accept plain": {
			giveHeaders: map[string][]string{"Accept": {"text/plaiN,text/html,application/xml;q=0.9,,,*/*;q=0.8"}},
			wantFormat:  plainTextFormat,
		},
		"accept json, weighted values only": {
			giveHeaders: map[string][]string{"Accept": {"application/jsoN;Q=0.1,text/html;q=1.1,application/xml;q=-1,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},

		"x-format json, depends on weight": {
			giveHeaders: map[string][]string{"X-Format": {"application/jsoN,*/*;q=0.8"}},
			wantFormat:  jsonFormat,
		},
		"x-format xml": {
			giveHeaders: map[string][]string{"X-Format": {"application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"}},
			wantFormat:  xmlFormat,
		},

		"content type has priority over accept": {
			giveHeaders: map[string][]string{"Content-Type": {"text/plain"}, "Accept": {"application/xml"}},
			wantFormat:  plainTextFormat,
		},
		"accept has priority over x-format": {
			giveHeaders: map[string][]string{"Accept": {"application/xml"}, "X-Format": {"text/plain"}},
			wantFormat:  plainTextFormat,
		},

		"empty headers": {
			giveHeaders: nil,
		},
		"empty content type": {
			giveHeaders: map[string][]string{"Content-Type": {"  "}},
		},
		"wrong content type": {
			giveHeaders: map[string][]string{"Content-Type": {"multipart/form-data; boundary=something"}},
		},
		"wrong accept": {
			giveHeaders: map[string][]string{"Accept": {";q=foobar,bar/baz;;;;;application/xml"}},
		},
		"none on invalid input": {
			giveHeaders: map[string][]string{"Content-Type": {"foo/bar; charset=utf-8"}, "Accept": {"foo/bar; charset=utf-8"}},
		},
		"completely unknown": {
			giveHeaders: map[string][]string{"Content-Type": {"😀"}, "Accept": {"😄"}, "X-Format": {"😍"}},
		},
	} {
		tt := _tt

		t.Run(name, func(t *testing.T) {
			var headers = new(fasthttp.RequestHeader)

			for key, values := range tt.giveHeaders {
				for _, value := range values {
					headers.Add(key, value)
				}
			}

			assert.Equal(t, tt.wantFormat, detectPreferredFormatForClient(headers))
		})
	}
}
