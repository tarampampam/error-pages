package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/config"
)

func TestRotationMode_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "disabled", config.RotationModeDisabled.String())
	assert.Equal(t, "random-on-startup", config.RotationModeRandomOnStartup.String())
	assert.Equal(t, "random-on-each-request", config.RotationModeRandomOnEachRequest.String())
	assert.Equal(t, "random-daily", config.RotationModeRandomDaily.String())
	assert.Equal(t, "random-hourly", config.RotationModeRandomHourly.String())

	assert.Equal(t, "RotationMode(255)", config.RotationMode(255).String())
}

func TestRotationModes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []config.RotationMode{
		config.RotationModeDisabled,
		config.RotationModeRandomOnStartup,
		config.RotationModeRandomOnEachRequest,
		config.RotationModeRandomDaily,
		config.RotationModeRandomHourly,
	}, config.RotationModes())
}

func TestRotationModeStrings(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{
		"disabled",
		"random-on-startup",
		"random-on-each-request",
		"random-daily",
		"random-hourly",
	}, config.RotationModeStrings())
}

func TestParseRotationMode(t *testing.T) {
	t.Parallel()

	for name, _tt := range map[string]struct {
		giveBytes    []byte
		giveString   string
		wantMode     config.RotationMode
		wantErrorMsg string
	}{
		"<empty string>":            {giveString: "", wantMode: config.RotationModeDisabled},
		"<empty bytes>":             {giveBytes: []byte(""), wantMode: config.RotationModeDisabled},
		"disabled":                  {giveString: "disabled", wantMode: config.RotationModeDisabled},
		"disabled (bytes)":          {giveBytes: []byte("disabled"), wantMode: config.RotationModeDisabled},
		"random-on-startup":         {giveString: "random-on-startup", wantMode: config.RotationModeRandomOnStartup},
		"random-on-startup (bytes)": {giveBytes: []byte("random-on-startup"), wantMode: config.RotationModeRandomOnStartup},
		"on-each-request":           {giveString: "random-on-each-request", wantMode: config.RotationModeRandomOnEachRequest},
		"daily":                     {giveString: "random-daily", wantMode: config.RotationModeRandomDaily},
		"hourly":                    {giveString: "random-hourly", wantMode: config.RotationModeRandomHourly},

		"foobar": {giveString: "foobar", wantErrorMsg: "unrecognized rotation mode: \"foobar\""},
	} {
		tt := _tt

		t.Run(name, func(t *testing.T) {
			var (
				mode config.RotationMode
				err  error
			)

			if tt.giveString != "" || tt.giveBytes == nil {
				mode, err = config.ParseRotationMode(tt.giveString)
			} else {
				mode, err = config.ParseRotationMode(tt.giveBytes)
			}

			if tt.wantErrorMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMode, mode)
			} else {
				assert.ErrorContains(t, err, tt.wantErrorMsg)
			}
		})
	}
}
