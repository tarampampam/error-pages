package logger_test

import (
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestFormat_String(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		give logger.Format
		want string
	}{
		"console":   {give: logger.ConsoleFormat, want: "console"},
		"json":      {give: logger.JSONFormat, want: "json"},
		"<unknown>": {give: logger.Format(255), want: "format(255)"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestParseFormat(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		give       string
		wantFormat logger.Format
		wantErrMsg string
	}{
		"<empty>": {give: "", wantFormat: logger.ConsoleFormat},
		"console": {give: "console", wantFormat: logger.ConsoleFormat},
		"json":    {give: "json", wantFormat: logger.JSONFormat},
		"foobar":  {give: "foobar", wantErrMsg: `unrecognized logging format: "foobar"`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := logger.ParseFormat(tt.give)

			if tt.wantErrMsg != "" {
				assert.ErrorEqual(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantFormat, got)
			}
		})
	}
}
