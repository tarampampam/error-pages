package tpl

import (
	"bytes"
	"encoding/json"
	"text/template"
)

// RenderHTML makes a replaces in the HTML-formatted content. Tokens for the replacement must be wrapped in double
// braces (with a single space or without it). Token examples:
//	{{ foo }}
//	{{bar}}
// Additionally, the Go template codes are fully supported in the template content.
func RenderHTML(content []byte, props Properties) ([]byte, error) {
	if len(content) == 0 {
		return content, nil
	}

	for what, with := range props.Replaces() {
		var n = []byte(with)

		content = bytes.ReplaceAll(content, []byte("{{"+what+"}}"), n)
		content = bytes.ReplaceAll(content, []byte("{{ "+what+" }}"), n)
	}

	t, err := template.New("").Parse(string(content))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err = t.Execute(&buf, props); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func RenderJSON(content []byte, props Properties) ([]byte, error) {
	if len(content) == 0 {
		return content, nil
	}

	for what, with := range props.Replaces() {
		n, err := json.Marshal(with) // escape characters
		if err != nil {
			return nil, err
		}

		if len(n) >= 2 {
			n = n[1 : len(n)-1] // truncate the first and last characters - quotes
		}

		content = bytes.ReplaceAll(content, []byte("{{"+what+"}}"), n)
		content = bytes.ReplaceAll(content, []byte("{{ "+what+" }}"), n)
	}

	return content, nil
}
