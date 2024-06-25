package template

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gh.tarampamp.am/error-pages/internal/appmeta"
)

var builtInFunctions = template.FuncMap{ //nolint:gochecknoglobals
	// current time:
	//	`{{ now.Unix }}`	// `1631610000`
	//	`{{ now.Hour }}:{{ now.Minute }}:{{ now.Second }}`	// `15:4:5`
	"now": time.Now,

	// current hostname:
	//	`{{ hostname }}`	// `localhost`
	"hostname": func() string { h, _ := os.Hostname(); return h }, //nolint:nlreturn

	// json-serialized value (safe to use with any type):
	//	`{{ json "test" }}`	// `"test"`
	//	`{{ json 42 }}`	// `42`
	"json": func(v any) string { b, _ := json.Marshal(v); return string(b) }, //nolint:nlreturn,errchkjson

	// cast any type to int, or return 0 if it's not possible:
	//	`{{ int "42" }}`	// `42`
	//	`{{ int 42 }}`	// `42`
	//	`{{ int 3.14 }}`	// `3`
	//	`{{ int "test" }}`	// `0`
	//	`{{ int "42test" }}`	// `0`
	"int": func(v any) int { // cast any type to int, or return 0 if it's not possible
		switch v := v.(type) {
		case string:
			if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				return i
			}
		case int:
			return v
		case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			if i, err := strconv.Atoi(fmt.Sprintf("%d", v)); err == nil { // not effective, but safe
				return i
			}
		case float32, float64:
			if i, err := strconv.ParseFloat(fmt.Sprintf("%f", v), 32); err == nil { // not effective, but safe
				return int(i)
			}
		case fmt.Stringer:
			if i, err := strconv.Atoi(v.String()); err == nil {
				return i
			}
		}

		return 0
	},

	// current application version:
	//	`{{ version }}`	// `1.0.0`
	"version": appmeta.Version,

	// counts the number of non-overlapping instances of substr in s:
	//	`{{ strCount "test" "t" }}`	// `2`
	"strCount": strings.Count,

	// reports whether substr is within s:
	//	`{{ strContains "test" "es" }}`	// `true`
	//	`{{ strContains "test" "ez" }}`	// `false`
	"strContains": strings.Contains,

	// returns a slice of the string s, with all leading and trailing white space removed:
	//	`{{ strTrimSpace "  test  " }}`	// `test`
	"strTrimSpace": strings.TrimSpace,

	// returns s without the provided leading prefix string:
	//	`{{ strTrimPrefix "test" "te" }}`	// `st`
	"strTrimPrefix": strings.TrimPrefix,

	// returns s without the provided trailing suffix string:
	//	`{{ strTrimSuffix "test" "st" }}`	// `te`
	"strTrimSuffix": strings.TrimSuffix,

	// returns a copy of the string s with all non-overlapping instances of old replaced by new:
	//	`{{ strReplace "test" "t" "z" }}`	// `zesz`
	"strReplace": strings.ReplaceAll,

	// returns the index of the first instance of substr in s, or -1 if substr is not present in s:
	//	`{{ strIndex "barfoobaz" "foo" }}`	// `3`
	"strIndex": strings.Index,

	// splits the string s around each instance of one or more consecutive white space characters:
	//	`{{ strFields "foo bar baz" }}`	// `[foo bar baz]`
	"strFields": strings.Fields,

	// retrieves the value of the environment variable named by the key:
	//	`{{ env "SHELL" }}`	// `/bin/bash`
	"env": os.Getenv,
}

func Render(content string, props Props) (string, error) {
	var fns = maps.Clone(builtInFunctions)

	maps.Copy(fns, template.FuncMap{ // add custom functions
		"hide_details": func() bool { return !props.ShowRequestDetails }, // inverted logic
		"l10n_enabled": func() bool { return !props.L10nDisabled },       // inverted logic
	})

	// allow the direct access to the properties tokens, e.g. `{{ service_port | json }}`
	// instead of `{{ .service_port | json }}`
	for k, v := range props.Values() {
		fns[k] = func() any { return v }
	}

	tmpl, tErr := template.New("template").Funcs(fns).Parse(content)
	if tErr != nil {
		return "", fmt.Errorf("failed to parse template: %w", tErr)
	}

	var buf strings.Builder

	if err := tmpl.Execute(&buf, props); err != nil {
		return "", err
	}

	return buf.String(), nil
}
