package tpl

import (
	"bytes"
	"io"
	"text/template"
)

// Template is a parsed error page template ready to be rendered with [Data].
type Template struct {
	tpl *template.Template
}

// New parses src as a Go template and returns a [Template] ready for rendering.
func New(src string) (*Template, error) {
	tpl, tErr := template.New("tpl").Funcs(fns).Parse(convertV3toV4(src))
	if tErr != nil {
		return nil, tErr
	}

	return &Template{tpl: tpl}, nil
}

// RenderTo executes the template with the given data and writes the result to dst.
func (t *Template) RenderTo(data Data, dst io.Writer) error { return t.tpl.Execute(dst, data) }

// Render executes the template with the given data and returns the result as a byte slice.
func (t *Template) Render(data Data) ([]byte, error) {
	var buf bytes.Buffer

	if err := t.RenderTo(data, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
