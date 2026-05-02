package logger_test

import (
	"bytes"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNewStdLog(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		giveLevel logger.Level
		wantLevel string
	}{
		"debug": {logger.DebugLevel, "debug"},
		"info":  {logger.InfoLevel, "info"},
		"warn":  {logger.WarnLevel, "warn"},
		"error": {logger.ErrorLevel, "error"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			log, err := logger.New(logger.DebugLevel, logger.JSONFormat, logger.WithWriter(&buf))

			assert.NoError(t, err)

			std := logger.NewStdLog(log, tt.giveLevel)
			std.Print("test message")

			assert.Contains(t, buf.String(), "test message")
			assert.Contains(t, buf.String(), tt.wantLevel)
		})
	}

	t.Run("named logger", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		log, err := logger.New(logger.DebugLevel, logger.JSONFormat, logger.WithWriter(&buf))

		assert.NoError(t, err)

		std := logger.NewStdLog(log.Named("http.server"), logger.WarnLevel)
		std.Print("listen error")

		assert.Contains(t, buf.String(), `"logger":"http.server"`)
		assert.Contains(t, buf.String(), "listen error")
	})
}
