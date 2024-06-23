package config_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		var cfg = config.New()

		assert.NotEmpty(t, cfg.Formats.XML)
		assert.NotEmpty(t, cfg.Formats.JSON)
		assert.True(t, len(cfg.Codes) >= 19)
		assert.True(t, len(cfg.Templates) >= 2)
		assert.NotEmpty(t, cfg.TemplateName)
		assert.True(t, cfg.Templates.Has(cfg.TemplateName))
		assert.Equal(t, uint16(http.StatusNotFound), cfg.DefaultCodeToRender)
	})

	t.Run("changing cfg1 should not affect cfg2", func(t *testing.T) {
		var cfg1, cfg2 = config.New(), config.New()

		cfg1.Codes["400"] = config.CodeDescription{Message: "foo", Description: "bar"}

		assert.NotEqual(t, cfg1.Codes["400"], cfg2.Codes["400"])

		cfg1.ProxyHeaders = append(cfg1.ProxyHeaders, "foo")

		assert.NotEqual(t, cfg1.ProxyHeaders, cfg2.ProxyHeaders)
	})
}
