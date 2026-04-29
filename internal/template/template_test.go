package tpl_test

import (
	"testing"

	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
	"gh.tarampamp.am/error-pages/v4/templates"
)

func TestTemplate_Render(t *testing.T) {
	t.Parallel()

	fullData := tpl.Data{
		StatusCode:   123,
		Message:      "Test Message",
		Description:  "Test Description",
		OriginalURI:  "/test",
		Namespace:    "test-namespace",
		IngressName:  "test-ingress",
		ServiceName:  "test-service",
		ServicePort:  "8080",
		RequestID:    "test-request-id",
		ForwardedFor: "123.123.123.123:321",
		Host:         "test-host",
		Config: tpl.Config{
			ShowRequestDetails: true,
			L10nDisabled:       true,
		},
	}

	t.Run("render built-in templates", func(t *testing.T) {
		t.Parallel()

		for name, content := range templates.BuiltInHTML() {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				t.Run("with data", func(t *testing.T) {
					t.Parallel()

					template, err := tpl.New(content)
					assert.NoError(t, err)

					_, err = template.Render(fullData)
					assert.NoError(t, err)
				})

				t.Run("without data", func(t *testing.T) {
					t.Parallel()

					template, err := tpl.New(content)
					assert.NoError(t, err)

					_, err = template.Render(tpl.Data{})
					assert.NoError(t, err)
				})
			})
		}

		t.Run("json, xml, plain text, etc", func(t *testing.T) {
			t.Parallel()

			for _, content := range []string{
				templates.JSON,
				templates.XML,
				templates.PlaintText,
			} {
				t.Run(content, func(t *testing.T) {
					t.Parallel()

					template, err := tpl.New(content)
					assert.NoError(t, err)

					_, err = template.Render(fullData)
					assert.NoError(t, err)
				})
			}
		})
	})
}
