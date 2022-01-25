package tpl

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/tarampampam/error-pages/internal/version"
)

var tplFnMap = template.FuncMap{ //nolint:gochecknoglobals // these functions can be used in templates
	"now":      time.Now,
	"hostname": os.Hostname,
	"json":     func(v interface{}) string { b, _ := json.Marshal(v); return string(b) }, //nolint:nlreturn
	"version":  version.Version,
	"int": func(v interface{}) int {
		if s, ok := v.(string); ok {
			if i, err := strconv.Atoi(s); err == nil {
				return i
			}
		} else if i, ok := v.(int); ok {
			return i
		}

		return 0
	},
}

func Render(content []byte, props Properties) ([]byte, error) {
	if len(content) == 0 {
		return content, nil
	}

	var funcMap = template.FuncMap{
		"show_details": func() bool { return props.ShowRequestDetails },
		"hide_details": func() bool { return !props.ShowRequestDetails },
	}

	// make a copy of template functions map
	for s, i := range tplFnMap {
		funcMap[s] = i
	}

	// and allow the direct calling of Properties tokens, e.g. `{{ code | json }}`
	for what, with := range props.Replaces() {
		var n, s = what, with

		funcMap[n] = func() string { return s }
	}

	t, err := template.New("").Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err = t.Execute(&buf, props); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
