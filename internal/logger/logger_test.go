package logger_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tarampampam/error-pages/internal/logger"
)

func TestNewDebugLevelConsoleFormat(t *testing.T) {
	output := capturer.CaptureStderr(func() {
		log, err := logger.New(logger.DebugLevel, logger.ConsoleFormat)
		require.NoError(t, err)

		log.Debug("dbg msg")
		log.Info("inf msg")
		log.Error("err msg")
	})

	assert.Contains(t, output, time.Now().Format("15:04:05"))
	assert.Regexp(t, `\t.+info.+\tinf msg`, output)
	assert.Regexp(t, `\t.+info.+\t.+logger_test\.go:\d+\tinf msg`, output)
	assert.Contains(t, output, "dbg msg")
	assert.Contains(t, output, "err msg")
}

func TestNewErrorLevelConsoleFormat(t *testing.T) {
	output := capturer.CaptureStderr(func() {
		log, err := logger.New(logger.ErrorLevel, logger.ConsoleFormat)
		require.NoError(t, err)

		log.Debug("dbg msg")
		log.Info("inf msg")
		log.Error("err msg")
	})

	assert.NotContains(t, output, "inf msg")
	assert.NotContains(t, output, "dbg msg")
	assert.Contains(t, output, "err msg")
}

func TestNewWarnLevelJSONFormat(t *testing.T) {
	output := capturer.CaptureStderr(func() {
		log, err := logger.New(logger.WarnLevel, logger.JSONFormat)
		require.NoError(t, err)

		log.Debug("dbg msg")
		log.Info("inf msg")
		log.Warn("warn msg")
		log.Error("err msg")
	})

	// replace timestamp field with fixed value
	fakeTimestamp := regexp.MustCompile(`"ts":\d+\.\d+,`)
	output = fakeTimestamp.ReplaceAllString(output, `"ts":0.1,`)

	lines := strings.Split(strings.Trim(output, "\n"), "\n")

	assert.JSONEq(t, `{"level":"warn","ts":0.1,"msg":"warn msg"}`, lines[0])
	assert.JSONEq(t, `{"level":"error","ts":0.1,"msg":"err msg"}`, lines[1])
}

func TestNewErrors(t *testing.T) {
	_, err := logger.New(logger.Level(127), logger.ConsoleFormat)
	require.EqualError(t, err, "unsupported logging level")

	_, err = logger.New(logger.WarnLevel, logger.Format(255))
	require.EqualError(t, err, "unsupported logging format")
}
