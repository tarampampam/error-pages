package template_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/appmeta"
	"gh.tarampamp.am/error-pages/internal/template"
	"gh.tarampamp.am/error-pages/l10n"
)

func TestRender_BuiltInFunction(t *testing.T) {
	t.Parallel()

	var hostname, hErr = os.Hostname()

	require.NoError(t, hErr)

	for name, tt := range map[string]struct {
		giveTemplate string
		wantResult   string
		wantErrMsg   string
	}{
		"now (unix)": {
			giveTemplate: `{{ nowUnix }}`,
			wantResult:   strconv.Itoa(int(time.Now().Unix())),
		},
		"hostname":                  {giveTemplate: `{{ hostname }}`, wantResult: hostname},
		"json (string)":             {giveTemplate: `{{ json "test" }}`, wantResult: `"test"`},
		"json (int)":                {giveTemplate: `{{ json 42 }}`, wantResult: `42`},
		"json (func result)":        {giveTemplate: `{{ json hostname }}`, wantResult: `"` + hostname + `"`},
		"int (string)":              {giveTemplate: `{{ int "42" }}`, wantResult: `42`},
		"int (int)":                 {giveTemplate: `{{ int 42 }}`, wantResult: `42`},
		"int (float)":               {giveTemplate: `{{ int 3.14 }}`, wantResult: `3`},
		"int (wrong string)":        {giveTemplate: `{{ int "test" }}`, wantResult: `0`},
		"int (string with numbers)": {giveTemplate: `{{ int "42test" }}`, wantResult: `0`},
		"version":                   {giveTemplate: `{{ version }}`, wantResult: appmeta.Version()},
		"strCount":                  {giveTemplate: `{{ strCount "test" "t" }}`, wantResult: `2`},
		"strContains (true)":        {giveTemplate: `{{ strContains "test" "es" }}`, wantResult: `true`},
		"strContains (false)":       {giveTemplate: `{{ strContains "test" "ez" }}`, wantResult: `false`},
		"strTrimSpace":              {giveTemplate: `{{ strTrimSpace "  test  " }}`, wantResult: `test`},
		"strTrimPrefix":             {giveTemplate: `{{ strTrimPrefix "test" "te" }}`, wantResult: `st`},
		"strTrimSuffix":             {giveTemplate: `{{ strTrimSuffix "test" "st" }}`, wantResult: `te`},
		"strReplace":                {giveTemplate: `{{ strReplace "test" "t" "z" }}`, wantResult: `zesz`},
		"strIndex":                  {giveTemplate: `{{ strIndex "barfoobaz" "foo" }}`, wantResult: `3`},
		"strFields":                 {giveTemplate: `{{ strFields "foo bar baz" }}`, wantResult: `[foo bar baz]`},
		"env (ok)":                  {giveTemplate: `{{ env "TEST_ENV_VAR" }}`, wantResult: "unit-test"},
		"env (not found)":           {giveTemplate: `{{ env "NOT_FOUND_ENV_VAR" }}`, wantResult: ""},
		"l10nScript":                {giveTemplate: `{{ l10nScript }}`, wantResult: l10n.L10n()},
		"escape": {
			giveTemplate: `{{ escape "<script>alert('XSS' + \"HERE\")</script>" }}`,
			wantResult:   "&lt;script&gt;alert(&#39;XSS&#39; + &#34;HERE&#34;)&lt;/script&gt;",
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.Setenv("TEST_ENV_VAR", "unit-test"))

			defer func() { require.NoError(t, os.Unsetenv("TEST_ENV_VAR")) }()

			var result, err = template.Render(tt.giveTemplate, template.Props{})

			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestRender(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		giveTemplate string
		giveProps    template.Props
		wantResult   string
		wantErrMsg   string
	}{
		"common case": {
			giveTemplate: "{{code}}: {{ message }} {{description}}",
			giveProps:    template.Props{Code: 404, Message: "Not found", Description: "Blah"},
			wantResult:   "404: Not found Blah",
		},
		"html markup": {
			giveTemplate: "<!-- comment --><html><body>{{code}}: {{ message }} {{description}}</body></html>",
			giveProps:    template.Props{Code: 201, Message: "lorem ipsum"},
			wantResult:   "<!-- comment --><html><body>201: lorem ipsum </body></html>",
		},
		"with line breakers": {
			giveTemplate: "\t {{code | json}}: {{ message }} {{description}}\n",
			giveProps:    template.Props{},
			wantResult:   "\t 0:  \n",
		},
		"golang template": {
			giveTemplate: "\t {{code}} {{ .Code }}{{ if .Message }} Yeah {{end}}",
			giveProps:    template.Props{Code: 201, Message: "lorem ipsum"},
			wantResult:   "\t 201 201 Yeah ",
		},

		"json common case": {
			giveTemplate: `{"code": {{code | json}}, "message": {"here":[ {{ message | json }} ]}, "desc": "{{description}}"}`,
			giveProps:    template.Props{Code: 404, Message: "'\"{Not found\t\r\n"},
			wantResult:   `{"code": 404, "message": {"here":[ "'\"{Not found\t\r\n" ]}, "desc": ""}`,
		},
		"json golang template": {
			giveTemplate: `{"code": "{{code}}", "message": {"here":[ "{{ if .Message }} Yeah {{end}}" ]}}`,
			giveProps:    template.Props{Code: 201, Message: "lorem ipsum"},
			wantResult:   `{"code": "201", "message": {"here":[ " Yeah " ]}}`,
		},

		"fn l10n_enabled": {
			giveTemplate: "{{ if l10n_enabled }}Y{{ else }}N{{ end }}",
			giveProps:    template.Props{L10nDisabled: true},
			wantResult:   "N",
		},
		"fn l10n_disabled": {
			giveTemplate: "{{ if l10n_disabled }}Y{{ else }}N{{ end }}",
			giveProps:    template.Props{L10nDisabled: true},
			wantResult:   "Y",
		},

		"complete example with every property and function": {
			giveProps: template.Props{
				Code:               404,
				Message:            "Not found",
				Description:        "Blah",
				OriginalURI:        "/test",
				Namespace:          "default",
				IngressName:        "test-ingress",
				ServiceName:        "test-service",
				ServicePort:        "80",
				RequestID:          "123456",
				ForwardedFor:       "123.123.123.123:321",
				Host:               "test-host",
				ShowRequestDetails: true,
				L10nDisabled:       false,
			},
			giveTemplate: `
				== Props as functions ==
				code: {{code}}
				message: {{message}}
				description: {{description}}
				original_uri: {{original_uri}}
				namespace: {{namespace}}
				ingress_name: {{ingress_name}}
				service_name: {{service_name}}
				service_port: {{service_port}}
				request_id: {{request_id}}
				forwarded_for: {{forwarded_for}}
				host: {{host}}
				show_details: {{show_details}}
				l10n_disabled: {{l10n_disabled}}

				== Props as properties ==
				.Code: {{ .Code }}
				.Message: {{ .Message }}
				.Description: {{ .Description }}
				.OriginalURI: {{ .OriginalURI }}
				.Namespace: {{ .Namespace }}
				.IngressName: {{ .IngressName }}
				.ServiceName: {{ .ServiceName }}
				.ServicePort: {{ .ServicePort }}
				.RequestID: {{ .RequestID }}
				.ForwardedFor: {{ .ForwardedFor }}
				.Host: {{ .Host }}
				.ShowRequestDetails: {{ .ShowRequestDetails }}
				.L10nDisabled: {{ .L10nDisabled }}

				== Custom functions ==
				hide_details: {{ hide_details }}
				l10n_enabled: {{ l10n_enabled }}
`,
			wantResult: `
				== Props as functions ==
				code: 404
				message: Not found
				description: Blah
				original_uri: /test
				namespace: default
				ingress_name: test-ingress
				service_name: test-service
				service_port: 80
				request_id: 123456
				forwarded_for: 123.123.123.123:321
				host: test-host
				show_details: true
				l10n_disabled: false

				== Props as properties ==
				.Code: 404
				.Message: Not found
				.Description: Blah
				.OriginalURI: /test
				.Namespace: default
				.IngressName: test-ingress
				.ServiceName: test-service
				.ServicePort: 80
				.RequestID: 123456
				.ForwardedFor: 123.123.123.123:321
				.Host: test-host
				.ShowRequestDetails: true
				.L10nDisabled: false

				== Custom functions ==
				hide_details: false
				l10n_enabled: true
`,
		},

		"wrong template":    {giveTemplate: `{{ foo() }}`, wantErrMsg: `function "foo" not defined`},
		"wrong template #2": {giveTemplate: `{{ fo`, wantErrMsg: "failed to parse template"},
	} {
		t.Run(name, func(t *testing.T) {
			var result, err = template.Render(tt.giveTemplate, tt.giveProps)

			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
