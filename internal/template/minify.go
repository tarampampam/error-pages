package template

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

var htmlMinify = func() *minify.M { //nolint:gochecknoglobals
	var m = minify.New()

	m.AddFunc("text/css", css.Minify)
	m.Add("text/html", &html.Minifier{KeepDocumentTags: true, KeepEndTags: true, KeepQuotes: true})
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFunc("application/javascript", js.Minify)

	return m
}()

// MiniHTML minifies HTML data, including inline CSS, SVG and JS.
func MiniHTML(data string) (string, error) { return htmlMinify.String("text/html", data) }
