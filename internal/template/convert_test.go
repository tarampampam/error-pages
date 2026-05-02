package tpl_test

import (
	"testing"

	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestConvertV3toV4(t *testing.T) {
	t.Parallel()

	t.Run("properties", func(t *testing.T) {
		t.Parallel()

		for name, tt := range map[string]struct {
			giveTemplate string
			giveData     tpl.Data
			wantResult   string
		}{
			"common case": {
				giveTemplate: "{{code}}: {{ message }} {{description}}",
				giveData:     tpl.Data{StatusCode: 404, Message: "Not found", Description: "Blah"},
				wantResult:   "404: Not found Blah",
			},
			"html markup": {
				giveTemplate: "<!-- comment --><html><body>{{code}}: {{ message }} {{description}}</body></html>",
				giveData:     tpl.Data{StatusCode: 201, Message: "lorem ipsum"},
				wantResult:   "<!-- comment --><html><body>201: lorem ipsum </body></html>",
			},
			"with line breakers": {
				giveTemplate: "\t {{code | json}}: {{ message }} {{description}}\n",
				giveData:     tpl.Data{},
				wantResult:   "\t 0:  \n",
			},
			"golang template": {
				giveTemplate: "\t {{code}} {{ .Code }}{{ if .Message }} Yeah {{end}}",
				giveData:     tpl.Data{StatusCode: 201, Message: "lorem ipsum"},
				wantResult:   "\t 201 201 Yeah ",
			},

			"json common case": {
				giveTemplate: `{"code": {{code | json}}, "message": {"here":[ {{ message | json }} ]}, "desc": "{{description}}"}`,
				giveData:     tpl.Data{StatusCode: 404, Message: "'\"{Not found\t\r\n"},
				wantResult:   `{"code": 404, "message": {"here":[ "'\"{Not found\t\r\n" ]}, "desc": ""}`,
			},
			"json golang template": {
				giveTemplate: `{"code": "{{code}}", "message": {"here":[ "{{ if .Message }} Yeah {{end}}" ]}}`,
				giveData:     tpl.Data{StatusCode: 201, Message: "lorem ipsum"},
				wantResult:   `{"code": "201", "message": {"here":[ " Yeah " ]}}`,
			},

			"already v4 .Config.ShowRequestDetails not double-converted": {
				giveTemplate: "{{ if .Config.ShowRequestDetails }}Y{{ else }}N{{ end }}",
				giveData:     tpl.Data{Config: tpl.Config{ShowRequestDetails: true}},
				wantResult:   "Y",
			},
			"already v4 .Config.L10nDisabled not double-converted": {
				giveTemplate: "{{ if .Config.L10nDisabled }}Y{{ else }}N{{ end }}",
				giveData:     tpl.Data{Config: tpl.Config{L10nDisabled: true}},
				wantResult:   "Y",
			},
			"fn l10n_enabled": {
				giveTemplate: "{{ if l10n_enabled }}Y{{ else }}N{{ end }}",
				giveData:     tpl.Data{Config: tpl.Config{L10nDisabled: true}},
				wantResult:   "N",
			},
			"fn l10n_disabled": {
				giveTemplate: "{{ if l10n_disabled }}Y{{ else }}N{{ end }}",
				giveData:     tpl.Data{Config: tpl.Config{L10nDisabled: true}},
				wantResult:   "Y",
			},

			"complete example with every property and function": {
				giveData: tpl.Data{
					StatusCode:   404,
					Message:      "Not found",
					Description:  "Blah",
					OriginalURI:  "/test",
					Namespace:    "default",
					IngressName:  "test-ingress",
					ServiceName:  "test-service",
					ServicePort:  "80",
					RequestID:    "123456",
					ForwardedFor: "123.123.123.123:321",
					Host:         "test-host",
					Config: tpl.Config{
						ShowRequestDetails: true,
						L10nDisabled:       false,
					},
				},
				giveTemplate: `
				== Props as functions ==
				code: {{code}}
				message: {{message}}
				description: {{description}}
				original_uri: {{original_uri}}
				namespace: {{namespace}}
				ingress_name: {{ ingress_name }}
				service_name: {{service_name}}
				service_port: {{ service_port}}
				request_id: {{request_id}}
				forwarded_for: {{forwarded_for }}
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
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				tmpl, err := tpl.New(tt.giveTemplate)
				assert.NoError(t, err)

				result, err := tmpl.Render(tt.giveData)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, string(result))
			})
		}
	})
}
