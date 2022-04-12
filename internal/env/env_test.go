package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, "LISTEN_ADDR", string(ListenAddr))
	assert.Equal(t, "LISTEN_PORT", string(ListenPort))
	assert.Equal(t, "TEMPLATE_NAME", string(TemplateName))
	assert.Equal(t, "CONFIG_FILE", string(ConfigFilePath))
	assert.Equal(t, "DEFAULT_ERROR_PAGE", string(DefaultErrorPage))
	assert.Equal(t, "DEFAULT_HTTP_CODE", string(DefaultHTTPCode))
	assert.Equal(t, "SHOW_DETAILS", string(ShowDetails))
	assert.Equal(t, "PROXY_HTTP_HEADERS", string(ProxyHTTPHeaders))
	assert.Equal(t, "DISABLE_L10N", string(DisableL10n))
}

func TestEnvVariable_Lookup(t *testing.T) {
	cases := []struct {
		giveEnv envVariable
	}{
		{giveEnv: ListenAddr},
		{giveEnv: ListenPort},
		{giveEnv: TemplateName},
		{giveEnv: ConfigFilePath},
		{giveEnv: DefaultErrorPage},
		{giveEnv: DefaultHTTPCode},
		{giveEnv: ShowDetails},
		{giveEnv: ProxyHTTPHeaders},
		{giveEnv: DisableL10n},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveEnv.String(), func(t *testing.T) {
			assert.NoError(t, os.Unsetenv(tt.giveEnv.String())) // make sure that env is unset for test

			defer func() { assert.NoError(t, os.Unsetenv(tt.giveEnv.String())) }()

			value, exists := tt.giveEnv.Lookup()
			assert.False(t, exists)
			assert.Empty(t, value)

			assert.NoError(t, os.Setenv(tt.giveEnv.String(), "foo"))

			value, exists = tt.giveEnv.Lookup()
			assert.True(t, exists)
			assert.Equal(t, "foo", value)
		})
	}
}
