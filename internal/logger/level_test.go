package logger_test

import (
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestLevel_String(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		give logger.Level
		want string
	}{
		"debug":     {give: logger.DebugLevel, want: "debug"},
		"info":      {give: logger.InfoLevel, want: "info"},
		"warn":      {give: logger.WarnLevel, want: "warn"},
		"error":     {give: logger.ErrorLevel, want: "error"},
		"<unknown>": {give: logger.Level(127), want: "level(127)"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		give       string
		wantLevel  logger.Level
		wantErrMsg string
	}{
		"<empty>": {give: "", wantLevel: logger.InfoLevel},
		"trace":   {give: "trace", wantLevel: logger.DebugLevel},
		"verbose": {give: "verbose", wantLevel: logger.DebugLevel},
		"debug":   {give: "debug", wantLevel: logger.DebugLevel},
		"info":    {give: "info", wantLevel: logger.InfoLevel},
		"warn":    {give: "warn", wantLevel: logger.WarnLevel},
		"warning": {give: "warning", wantLevel: logger.WarnLevel},
		"err":     {give: "err", wantLevel: logger.ErrorLevel},
		"error":   {give: "error", wantLevel: logger.ErrorLevel},
		"foobar":  {give: "foobar", wantErrMsg: `unrecognized logging level: "foobar"`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := logger.ParseLevel(tt.give)

			if tt.wantErrMsg != "" {
				assert.ErrorEqual(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLevel, got)
			}
		})
	}
}
