package formats_test

import (
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/formats"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestFormat_ContentType(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		give formats.Format
		want string
	}{
		"plain text": {give: formats.PlainTextFormat, want: "text/plain; charset=utf-8"},
		"html":       {give: formats.HTMLFormat, want: "text/html; charset=utf-8"},
		"json":       {give: formats.JSONFormat, want: "application/json; charset=utf-8"},
		"xml":        {give: formats.XMLFormat, want: "application/xml; charset=utf-8"},
		"unknown":    {give: formats.Format(255), want: ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.give.ContentType())
		})
	}
}

func TestFormat_FormatError(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		giveFormat formats.Format
		giveErr    string
		want       string
	}{
		"plain/simple": {
			giveFormat: formats.PlainTextFormat,
			giveErr:    "page not found",
			want:       "page not found",
		},
		"plain/empty": {
			giveFormat: formats.PlainTextFormat,
			giveErr:    "",
			want:       "",
		},
		"plain/special chars passthrough": {
			giveFormat: formats.PlainTextFormat,
			giveErr:    `<b>error</b> & "more"`,
			want:       `<b>error</b> & "more"`,
		},
		"json/simple": {
			giveFormat: formats.JSONFormat,
			giveErr:    "page not found",
			want:       `{"error":"page not found"}`,
		},
		"json/empty": {
			giveFormat: formats.JSONFormat,
			giveErr:    "",
			want:       `{"error":""}`,
		},
		"json/html chars escaped": {
			// Go's json.Marshal escapes <, >, & to \uXXXX sequences to prevent XSS when JSON is embedded in HTML
			giveFormat: formats.JSONFormat,
			giveErr:    `<b>bold</b> & "quoted"`,
			want:       "{\"error\":\"\\u003cb\\u003ebold\\u003c/b\\u003e \\u0026 \\\"quoted\\\"\"}",
		},
		"xml/simple": {
			giveFormat: formats.XMLFormat,
			giveErr:    "page not found",
			want:       "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<error>page not found</error>",
		},
		"xml/empty": {
			giveFormat: formats.XMLFormat,
			giveErr:    "",
			want:       "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<error></error>",
		},
		"xml/special chars escaped": {
			giveFormat: formats.XMLFormat,
			giveErr:    `<b>error</b> & 'it'`,
			want:       "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<error>&lt;b&gt;error&lt;/b&gt; &amp; &#39;it&#39;</error>",
		},
		"html/simple": {
			giveFormat: formats.HTMLFormat,
			giveErr:    "page not found",
			want:       "<!DOCTYPE html>\n<html><head><meta charset=\"UTF-8\"></head><body>\npage not found\n</body></html>",
		},
		"html/empty": {
			giveFormat: formats.HTMLFormat,
			giveErr:    "",
			want:       "<!DOCTYPE html>\n<html><head><meta charset=\"UTF-8\"></head><body>\n\n</body></html>",
		},
		"html/special chars escaped": {
			giveFormat: formats.HTMLFormat,
			giveErr:    `<script>alert("xss")</script>`,
			want: "<!DOCTYPE html>\n<html><head><meta charset=\"UTF-8\"></head><body>\n" +
				"&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;\n</body></html>",
		},
		"unknown format": {
			giveFormat: formats.Format(255),
			giveErr:    "something went wrong",
			want:       "something went wrong",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, string(tt.giveFormat.FormatError(tt.giveErr)))
		})
	}
}
