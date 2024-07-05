package config_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/template"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		var cfg = config.New()

		assert.NotEmpty(t, cfg.Formats.XML)
		assert.NotEmpty(t, cfg.Formats.JSON)
		assert.NotEmpty(t, cfg.Formats.PlainText)
		assert.True(t, len(cfg.Codes) >= 19)
		assert.True(t, len(cfg.Templates) >= 1)
		assert.NotEmpty(t, cfg.TemplateName)
		assert.True(t, cfg.Templates.Has(cfg.TemplateName))
		assert.Equal(t, uint16(http.StatusNotFound), cfg.DefaultCodeToRender)
		assert.False(t, cfg.DisableMinification)
	})

	t.Run("changing cfg1 should not affect cfg2", func(t *testing.T) {
		var cfg1, cfg2 = config.New(), config.New()

		cfg1.Codes["400"] = config.CodeDescription{Message: "foo", Description: "bar"}

		assert.NotEqual(t, cfg1.Codes["400"], cfg2.Codes["400"])

		cfg1.ProxyHeaders = append(cfg1.ProxyHeaders, "foo")

		assert.NotEqual(t, cfg1.ProxyHeaders, cfg2.ProxyHeaders)
	})

	t.Run("render default format templates", func(t *testing.T) {
		var cfg = config.New()

		for _, content := range []string{cfg.Formats.JSON, cfg.Formats.XML, cfg.Formats.PlainText} {
			var result, err = template.Render(content, template.Props{
				ShowRequestDetails: true,
				Code:               404,
				Message:            "Not Found",
			})

			assert.NotEmpty(t, result)
			assert.NoError(t, err)

			t.Log(result)
		}
	})
}
