package tpl_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
	"gh.tarampamp.am/error-pages/v4/l10n"
)

func TestFunctions(t *testing.T) {
	t.Parallel()

	t.Run("functions", func(t *testing.T) {
		t.Parallel()

		hostname, hErr := os.Hostname()
		assert.NoError(t, hErr)

		assert.NoError(t, os.Setenv("TEST_ENV_VAR", "unit-test")) //nolint:usetesting
		t.Cleanup(func() { assert.NoError(t, os.Unsetenv("TEST_ENV_VAR")) })

		assert.NoError(t, os.Setenv("SECRET_ENV_VAR", "supersecret")) //nolint:usetesting
		t.Cleanup(func() { assert.NoError(t, os.Unsetenv("SECRET_ENV_VAR")) })

		// time-sensitive: tested separately to avoid a second-boundary race between
		// map initialization and template rendering. string comparison is used to avoid
		// timezone mismatch that time.Parse introduces when no zone is in the layout.
		t.Run("now", func(t *testing.T) {
			t.Parallel()

			const layout = "2006-01-02 15:04:05"

			before := time.Now().Format(layout)

			tmpl, tplErr := tpl.New(`{{ now.Format "2006-01-02 15:04:05" }}`)
			assert.NoError(t, tplErr)

			result, renderErr := tmpl.Render(tpl.Data{})
			assert.NoError(t, renderErr)

			after := time.Now().Format(layout)

			s := string(result)
			assert.True(t, s == before || s == after)
		})

		for name, tt := range map[string]struct{ give, want string }{
			"hostname": {give: "{{ hostname }}", want: hostname},

			"toJson (string) pipe":      {give: `{{ "test" | toJson }}`, want: `"test"`},
			"toJson (string)":           {give: `{{ toJson "test" }}`, want: `"test"`},
			"toJson (int) pipe":         {give: "{{ 42 | toJson }}", want: "42"},
			"toJson (int)":              {give: "{{ toJson 42 }}", want: "42"},
			"toJSON (func result) pipe": {give: "{{ hostname | toJSON }}", want: `"` + hostname + `"`},
			"toJson (func result)":      {give: "{{ toJson hostname }}", want: `"` + hostname + `"`},

			"int (string)":              {give: `{{ int "42" }}`, want: "42"},
			"int (int) pipe":            {give: "{{42 | int | toJson}}", want: "42"},
			"int (float)":               {give: "{{ int 3.14 }}", want: "3"},
			"toInt (wrong string) pipe": {give: `{{ "test" | toInt }}`, want: "0"},
			"int (string with numbers)": {give: `{{ int "42test" }}`, want: "0"},
			"int (bool true)":           {give: `{{ true | int }}`, want: "1"},
			"int (bool false)":          {give: `{{ false | int }}`, want: "0"},
			"int (nil)":                 {give: `{{ int nil }}`, want: "0"},
			"int (string float)":        {give: `{{ int "3.14" }}`, want: "3"},
			"int (string spaces)":       {give: `{{ int "  42  " }}`, want: "42"},

			"version": {give: `{{ version }}`, want: appmeta.Version()},

			"env (ok)":        {give: `{{ env "TEST_ENV_VAR" }}`, want: "unit-test"},
			"env (not found)": {give: `{{ env "NOT_FOUND_ENV_VAR" }}`, want: ""},
			"env (secret)":    {give: `{{ env "SECRET_ENV_VAR" }}`, want: "***********"},

			"escape": {
				give: `{{ escape "<script>alert('XSS' + \"HERE\")</script>" }}`,
				want: "&lt;script&gt;alert(&#39;XSS&#39; + &#34;HERE&#34;)&lt;/script&gt;",
			},

			"trimPrefix":            {give: `{{ "test" | trimPrefix "te" }}`, want: "st"},
			"trimSuffix":            {give: `{{ "test" | trimSuffix "st" }}`, want: "te"},
			"trimSuffix (non-pipe)": {give: `{{ trimSuffix "st" "test"  }}`, want: "te"},
			"trimPostfix":           {give: `{{ "test" | trimPostfix "st" }}`, want: "te"},

			"replace": {give: `{{ "test" | replace "t" "z" }}`, want: "zesz"},

			"contains":     {give: `{{ "test" | contains "es" }}`, want: "true"},
			"not contains": {give: `{{ "test" | contains "z" }}`, want: "false"},

			"count":          {give: `{{ "test" | count "t" }}`, want: "2"},
			"count (substr)": {give: `{{ "testtest" | count "te" }}`, want: "2"},
			"count (zero)":   {give: `{{ "test" | count "z" }}`, want: "0"},

			"fields":         {give: `{{ "foo bar baz" | fields }}`, want: "[foo bar baz]"},
			"fields (empty)": {give: `{{ "" | fields }}`, want: "[]"},

			"lower": {give: `{{ "TEST" | lower }}`, want: "test"},
			"upper": {give: `{{ "test" | upper }}`, want: "TEST"},

			"default (env)":              {give: `{{ env "__TEST_NOT_SET" | default "def-value" }}`, want: "def-value"},
			"default (non-empty env)":    {give: `{{ env "TEST_ENV_VAR" | default "default-value" }}`, want: "unit-test"},
			"default (empty string)":     {give: `{{ "" | default "default-value" }}`, want: "default-value"},
			"default (non-empty string)": {give: `{{ "test" | default "default-value" }}`, want: "test"},
			"default (from data)":        {give: `{{ .OriginalURI | default "N/A" }}`, want: "N/A"},

			"hasPrefix":  {give: `{{ "test" | hasPrefix "te" }}`, want: "true"},
			"hasSuffix":  {give: `{{ "test" | hasSuffix "st" }}`, want: "true"},
			"hasPostfix": {give: `{{ "test" | hasPostfix "st" }}`, want: "true"},

			"split":              {give: `{{ "a,b,c" | split "," }}`, want: "[a b c]"},
			"split (single)":     {give: `{{ "abc" | split "," }}`, want: "[abc]"},
			"split then join":    {give: `{{ "a,b,c" | split "," | join " - " }}`, want: "a - b - c"},
			"join (from fields)": {give: `{{ "foo bar baz" | fields | join "-" }}`, want: "foo-bar-baz"},
			"join (non-pipe)":    {give: `{{ join ", " (split "," "a,b,c") }}`, want: "a, b, c"},
			"join (non-slice)":   {give: `{{ 42 | join "," }}`, want: "42"},

			"quote":                   {give: `{{ "test" | quote }}`, want: `"test"`},
			"quote (single quotes)":   {give: `{{ "it's here" | quote }}`, want: `"it's here"`},
			"quote (escapes newline)": {give: `{{ "test\n" | quote }}`, want: `"test\n"`},

			"squote":              {give: `{{ "test" | squote }}`, want: "'test'"},
			"squote (with space)": {give: `{{ "hello world" | squote }}`, want: "'hello world'"},

			"repeat": {give: `{{ "Ha" | repeat 3 }}`, want: "HaHaHa"},

			"toString (string)": {give: `{{ "test" | toString }}`, want: "test"},
			"toString (int)":    {give: `{{ 42 | toString }}`, want: "42"},
			"toString (float)":  {give: `{{ 3.14 | toString }}`, want: "3.14"},
			"toString (bool)":   {give: `{{ true | toString }}`, want: "true"},
			"str (alias)":       {give: `{{ 42 | str }}`, want: "42"},

			"ternary (true)":       {give: `{{ ternary "yes" "no" true }}`, want: "yes"},
			"ternary (false)":      {give: `{{ ternary "yes" "no" false }}`, want: "no"},
			"ternary (pipe true)":  {give: `{{ true | ternary "yes" "no" }}`, want: "yes"},
			"ternary (pipe false)": {give: `{{ false | ternary "yes" "no" }}`, want: "no"},

			"coalesce (first set)":  {give: `{{ coalesce "a" "b" }}`, want: "a"},
			"coalesce (skip empty)": {give: `{{ coalesce "" "b" }}`, want: "b"},
			"coalesce (skip two)":   {give: `{{ coalesce "" "" "c" }}`, want: "c"},
			"coalesce (all empty)":  {give: `{{ coalesce "" "" }}`, want: ""},

			"urlEncode (path)":   {give: `{{ "/api/v1" | urlEncode }}`, want: "%2Fapi%2Fv1"},
			"urlEncode (spaces)": {give: `{{ "hello world" | urlEncode }}`, want: "hello+world"},

			"isEmpty (empty string)":     {give: `{{ isEmpty "" }}`, want: "true"},
			"isEmpty (non-empty string)": {give: `{{ isEmpty "test" }}`, want: "false"},
			"isEmpty (zero int)":         {give: `{{ isEmpty 0 }}`, want: "true"},
			"isEmpty (nil)":              {give: `{{ isEmpty nil }}`, want: "true"},
			"isEmpty (bool false)":       {give: `{{ isEmpty false }}`, want: "true"},
			"isEmpty (bool true)":        {give: `{{ isEmpty true }}`, want: "false"},
			"isEmpty (float64 zero)":     {give: `{{ isEmpty 0.0 }}`, want: "true"},
			"isEmpty (float64 nonzero)":  {give: `{{ isEmpty 1.5 }}`, want: "false"},
			"isNotEmpty (empty string)":  {give: `{{ isNotEmpty "" }}`, want: "false"},
			"isNotEmpty (non-empty)":     {give: `{{ isNotEmpty "test" }}`, want: "true"},

			"truncate (short)":      {give: `{{ "test" | truncate 10 }}`, want: "test"},
			"truncate (exact)":      {give: `{{ "test" | truncate 4 }}`, want: "test"},
			"truncate (long)":       {give: `{{ "Hello, World!" | truncate 8 }}`, want: "Hello..."},
			"truncate (n=3)":        {give: `{{ "hello" | truncate 3 }}`, want: "..."},
			"truncate (n<ellipsis)": {give: `{{ "hello" | truncate 2 }}`, want: "he"},
			"truncate (n=0)":        {give: `{{ "hello" | truncate 0 }}`, want: ""},

			"trimAll":              {give: `{{ ".....test....." | trimAll "." }}`, want: "test"},
			"trimAll (multi-char)": {give: `{{ "!?test?!" | trimAll "!?" }}`, want: "test"},

			"substr":                         {give: `{{ "test" | substr 1 4 }}`, want: "est"},
			"substr (exact)":                 {give: `{{ "Hello, World!" | substr 7 5 }}`, want: "World"},
			"substr (negative start)":        {give: `{{ "test" | substr -1 2 }}`, want: "te"},
			"substr (negative length)":       {give: `{{ "test" | substr 2 -1 }}`, want: "st"},
			"substr (beyond length)":         {give: `{{ "test" | substr 1 100 }}`, want: "est"},
			"substr (start at end)":          {give: `{{ "test" | substr 4 1 }}`, want: ""},
			"substr (cyrillic)":              {give: `{{ "Привет" | substr 1 4 }}`, want: "риве"},
			"substr (cyrillic full)":         {give: `{{ "Привет" | substr 0 6 }}`, want: "Привет"},
			"substr (cyrillic negative len)": {give: `{{ "Привет" | substr 2 -1 }}`, want: "ивет"},
			"substr (emoji)":                 {give: `{{ "😊🔥🎉" | substr 1 2 }}`, want: "🔥🎉"},
			"substr (emoji single)":          {give: `{{ "😊🔥🎉" | substr 0 1 }}`, want: "😊"},

			"l10nScript": {give: `{{ l10nScript }}`, want: l10n.L10n()},
		} {
			t.Run(name, func(t *testing.T) {
				tmpl, err := tpl.New(tt.give)
				assert.NoError(t, err)

				result, err := tmpl.Render(tpl.Data{})
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(result))
			})
		}
	})

	// TODO: remove this test after removing deprecated functions from the codebase
	t.Run("deprecated functions are still available", func(t *testing.T) {
		t.Parallel()

		// time-sensitive: tested separately to avoid a second-boundary race
		t.Run("now (unix)", func(t *testing.T) {
			t.Parallel()

			before := time.Now().Unix()

			tmpl, tplErr := tpl.New("{{ nowUnix }}")
			assert.NoError(t, tplErr)

			result, renderErr := tmpl.Render(tpl.Data{})
			assert.NoError(t, renderErr)

			after := time.Now().Unix()

			got, convErr := strconv.ParseInt(string(result), 10, 64)
			assert.NoError(t, convErr)
			assert.True(t, got >= before && got <= after)
		})

		for name, tt := range map[string]struct{ give, want string }{
			"strCount":            {give: `{{ strCount "test" "t" }}`, want: "2"},
			"strContains (true)":  {give: `{{ strContains "test" "es" }}`, want: "true"},
			"strContains (false)": {give: `{{ strContains "test" "ez" }}`, want: "false"},
			"strTrimSpace":        {give: `{{strTrimSpace "  test  "}}`, want: "test"},
			"strTrimPrefix":       {give: `{{ strTrimPrefix "test" "te" }}`, want: "st"},
			"strTrimSuffix":       {give: `{{ strTrimSuffix "test" "st" }}`, want: "te"},
			"strReplace":          {give: `{{ strReplace "test" "t" "z" }}`, want: "zesz"},
			"strIndex":            {give: `{{ strIndex "barfoobaz" "foo" }}`, want: "3"},
			"strFields":           {give: `{{ strFields "foo bar baz" }}`, want: "[foo bar baz]"},

			"json (string) pipe": {give: `{{ "test" | json }}`, want: `"test"`},
			"json (string)":      {give: `{{ json "test" }}`, want: `"test"`},
			"json (int) pipe":    {give: "{{ 42 | json }}", want: "42"},
			"json (int)":         {give: "{{ json 42 }}", want: "42"},
		} {
			t.Run(name, func(t *testing.T) {
				tmpl, err := tpl.New(tt.give)
				assert.NoError(t, err)

				result, err := tmpl.Render(tpl.Data{})
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(result))
			})
		}
	})

	t.Run("functions with data", func(t *testing.T) {
		t.Parallel()

		for name, tt := range map[string]struct {
			give     string
			giveData tpl.Data
			want     string
		}{
			"toInt (uint16 field)":      {give: `{{ .StatusCode | int }}`, giveData: tpl.Data{StatusCode: 404}, want: "404"},
			"toString (uint16 field)":   {give: `{{ .StatusCode | toString }}`, giveData: tpl.Data{StatusCode: 200}, want: "200"},
			"isEmpty (uint16 zero)":     {give: `{{ .StatusCode | isEmpty }}`, giveData: tpl.Data{StatusCode: 0}, want: "true"},
			"isEmpty (uint16 non-zero)": {give: `{{ .StatusCode | isEmpty }}`, giveData: tpl.Data{StatusCode: 1}, want: "false"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				tmpl, err := tpl.New(tt.give)
				assert.NoError(t, err)

				result, err := tmpl.Render(tt.giveData)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(result))
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		for name, tt := range map[string]struct{ give, wantErrSubstr string }{
			"undefined function": {give: "{{ foo() }}", wantErrSubstr: `function "foo" not defined`},
			"invalid template":   {give: "{{- .fo", wantErrSubstr: "unclosed action"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				tmpl, err := tpl.New(tt.give)
				assert.Equal(t, nil, tmpl)
				assert.ErrorContains(t, err, tt.wantErrSubstr)
			})
		}
	})

	t.Run("render errors", func(t *testing.T) {
		t.Parallel()

		// "" | call tries to invoke a string as a function — parses OK, fails at execution.
		tmpl, err := tpl.New(`{{ "" | call }}`)
		assert.NoError(t, err)

		result, err := tmpl.Render(tpl.Data{})
		if result != nil {
			t.Errorf("expected nil result, got %q", result)
		}

		assert.ErrorContains(t, err, "non-function")
	})
}
