//go:build ignore

package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"slices"
	"strings"
	"text/template"
)

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"camel": func(filename string) string {
		parts := strings.FieldsFunc(filename, func(r rune) bool {
			return r == '-' || r == '_' || r == ' ' || r == '.' // split by common delimiters
		})

		var sb strings.Builder

		for _, p := range parts {
			if len(p) > 0 {
				sb.WriteString(strings.ToUpper(p[:1]) + p[1:])
			}
		}

		return sb.String()
	},
	"noExt": func(filename string) string {
		if i := strings.LastIndex(filename, "."); i >= 0 {
			return filename[:i]
		}

		return filename
	},
}).Parse(`// Code generated. DO NOT EDIT.

package templates

import _ "embed"

// Template content is loaded from the corresponding HTML files at compile time via go:embed.
var (
{{- range .FileNames }}
	//go:embed {{ . }}
	tpl{{ noExt . | camel }} string
{{ end }}
)

// TemplateName* constants hold the canonical name of each built-in template (filename without extension).
const (
{{- range .FileNames }}
	TemplateName{{ noExt . | camel }} = "{{ noExt . }}"
{{- end }}
)

// BuiltIn returns a new map of all built-in templates keyed by their canonical name.
// The map itself is freshly allocated on each call, so adding or removing keys is safe.
func BuiltIn() map[string]string {
	return map[string]string{
		{{- range .FileNames }}
		TemplateName{{ noExt . | camel }}: tpl{{ noExt . | camel }},
		{{- end }}
	}
}
`))

func main() {
	files, rErr := os.ReadDir(".")
	exitIfErr(rErr)

	var fileNames = make([]string, 0, len(files))

	for _, file := range files {
		if name := file.Name(); file.IsDir() ||
			(!strings.HasSuffix(name, ".html") && !strings.HasSuffix(name, ".htm")) ||
			strings.HasPrefix(name, ".") {
			continue // skip non-HTML files, hidden files, and directories
		}

		if info, infoErr := file.Info(); infoErr != nil {
			exitIfErr(infoErr)
		} else if info.Size() == 0 {
			continue // skip empty files
		}

		fileNames = append(fileNames, file.Name())
	}

	if len(fileNames) == 0 {
		cwd, _ := os.Getwd()

		exitIfErr(errors.New("no HTML files found in " + cwd))
	}

	slices.Sort(fileNames) // sort file names for deterministic output

	var buf bytes.Buffer
	exitIfErr(tmpl.Execute(&buf, struct {
		FileNames []string
	}{
		FileNames: fileNames,
	}))

	src, fmtErr := format.Source(buf.Bytes())
	if fmtErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, buf.String()) // print unformatted source for debugging
		exitIfErr(fmtErr)
	}

	exitIfErr(os.WriteFile("./embed.go", src, 0o644))
}

// exitIfErr prints the error message to stderr and exits with status code 1 if err is not nil.
func exitIfErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
