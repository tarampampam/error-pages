package logger_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/logger"
)

func TestNewNotVerboseDebugJSON(t *testing.T) {
	output := capturer.CaptureStderr(func() {
		log, err := logger.New(false, false, false)
		assert.NoError(t, err)

		log.Info("inf msg")
		log.Debug("dbg msg")
		log.Error("err msg")
	})

	assert.Contains(t, output, time.Now().Format("15:04:05"))
	assert.Regexp(t, `\t.+info.+\tinf msg`, output)
	assert.NotContains(t, output, "dbg msg")
	assert.Contains(t, output, "err msg")
}

func TestNewVerboseNotDebugJSON(t *testing.T) {
	output := capturer.CaptureStderr(func() {
		log, err := logger.New(true, false, false)
		assert.NoError(t, err)

		log.Info("inf msg")
		log.Debug("dbg msg")
		log.Error("err msg")
	})

	assert.Contains(t, output, time.Now().Format("15:04:05"))
	assert.Regexp(t, `\t.+info.+\tinf msg`, output)
	assert.Contains(t, output, "dbg msg")
	assert.Contains(t, output, "err msg")
}

func TestNewVerboseDebugNotJSON(t *testing.T) {
	output := capturer.CaptureStderr(func() {
		log, err := logger.New(true, true, false)
		assert.NoError(t, err)

		log.Info("inf msg")
		log.Debug("dbg msg")
		log.Error("err msg")
	})

	assert.Contains(t, output, time.Now().Format("15:04:05"))
	assert.Regexp(t, `\t.+info.+\t.+logger_test\.go:\d+\tinf msg`, output)
	assert.Contains(t, output, "dbg msg")
	assert.Contains(t, output, "err msg")
}

func TestNewNotVerboseDebugButJSON(t *testing.T) {
	output := capturer.CaptureStderr(func() {
		log, err := logger.New(false, false, true)
		assert.NoError(t, err)

		log.Info("inf msg")
		log.Debug("dbg msg")
		log.Error("err msg")
	})

	// replace timestamp field with fixed value
	fakeTimestamp := regexp.MustCompile(`"ts":\d+\.\d+,`)
	output = fakeTimestamp.ReplaceAllString(output, `"ts":0.1,`)

	lines := strings.Split(strings.Trim(output, "\n"), "\n")

	assert.JSONEq(t, `{"level":"info","ts":0.1,"msg":"inf msg"}`, lines[0])
	assert.JSONEq(t, `{"level":"error","ts":0.1,"msg":"err msg"}`, lines[1])
}
