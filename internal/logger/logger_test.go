package logger_test

import (
	"bytes"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestNewErrors(t *testing.T) {
	log, err := logger.New(logger.Level(127), logger.ConsoleFormat)
	require.Nil(t, log)
	require.EqualError(t, err, "unsupported logging level")

	log, err = logger.New(logger.WarnLevel, logger.Format(255))
	require.Nil(t, log)
	require.EqualError(t, err, "unsupported logging format")
}

func TestLogger_ConsoleFormat(t *testing.T) {
	var (
		buf         bytes.Buffer
		log, logErr = logger.New(logger.DebugLevel, logger.ConsoleFormat, &buf)

		now = time.Now()
	)

	require.NoError(t, logErr)
	assert.Equal(t, logger.DebugLevel, log.Level())

	log.Debug("debug message",
		logger.String("String", "value"),
		logger.Strings("Strings", "foo", "bar", ""),
		logger.Int64("Int64", 0),
		logger.Int("Int", 1),
		logger.Uint64("Uint64", 2),
		logger.Float64("Float64", 3.14),
		logger.Bool("Bool", true),
		logger.Time("Time", now),
		logger.Duration("Duration", time.Millisecond),
	)

	var output = buf.String()

	assert.Contains(t, output, `time=`+now.Format("15:04:")) // match without seconds
	assert.Contains(t, output, `level=debug`)
	assert.Contains(t, output, `msg="debug message"`)
	assert.Contains(t, output, "String=value")
	assert.Contains(t, output, `Strings="[foo bar ]"`)
	assert.Contains(t, output, "Int64=0")
	assert.Contains(t, output, "Int=1")
	assert.Contains(t, output, "Uint64=2")
	assert.Contains(t, output, "Float64=3.14")
	assert.Contains(t, output, "Bool=true")
	assert.Contains(t, output, "Time="+now.Format("2006-01-02T15:04:05.000Z07:00"))
	assert.Contains(t, output, "Duration=1ms")
}

func TestLogger_JSONFormat(t *testing.T) {
	var (
		buf         bytes.Buffer
		log, logErr = logger.New(logger.DebugLevel, logger.JSONFormat, &buf)

		now = time.Now()
	)

	require.NoError(t, logErr)
	assert.Equal(t, logger.DebugLevel, log.Level())

	log.Debug("debug message",
		logger.String("String", "value"),
		logger.Strings("Strings", "foo", "bar", ""),
		logger.Int64("Int64", 0),
		logger.Int("Int", 1),
		logger.Uint64("Uint64", 2),
		logger.Float64("Float64", 3.14),
		logger.Bool("Bool", true),
		logger.Time("Time", now),
		logger.Duration("Duration", time.Millisecond),
	)

	var output = buf.String()

	assert.Contains(t, output, `"ts":`+strconv.Itoa(int(now.Unix()))+".") // match without nanoseconds
	assert.Contains(t, output, `"level":"debug"`)
	assert.Contains(t, output, `"msg":"debug message"`)
	assert.Contains(t, output, `"String":"value"`)
	assert.Contains(t, output, `"Strings":["foo","bar",""]`)
	assert.Contains(t, output, `"Int64":0`)
	assert.Contains(t, output, `"Int":1`)
	assert.Contains(t, output, `"Uint64":2`)
	assert.Contains(t, output, `"Float64":3.14`)
	assert.Contains(t, output, `"Bool":true`)
	assert.Contains(t, output, `"Time":"`+now.Format("2006-01-02T15:04:05.000")) // omit nano seconds
	assert.Contains(t, output, `"Duration":1000000`)
}

func TestLogger_Debug(t *testing.T) {
	var (
		buf         bytes.Buffer
		log, logErr = logger.New(logger.DebugLevel, logger.JSONFormat, &buf)
	)

	require.NoError(t, logErr)
	assert.Equal(t, logger.DebugLevel, log.Level())

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	var output = buf.String()

	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestLogger_Info(t *testing.T) {
	var (
		buf         bytes.Buffer
		log, logErr = logger.New(logger.InfoLevel, logger.JSONFormat, &buf)
	)

	require.NoError(t, logErr)
	assert.Equal(t, logger.InfoLevel, log.Level())

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	var output = buf.String()

	assert.NotContains(t, output, "debug message")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestLogger_Warn(t *testing.T) {
	var (
		buf         bytes.Buffer
		log, logErr = logger.New(logger.WarnLevel, logger.JSONFormat, &buf)
	)

	require.NoError(t, logErr)
	assert.Equal(t, logger.WarnLevel, log.Level())

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	var output = buf.String()

	assert.NotContains(t, output, "debug message")
	assert.NotContains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestLogger_Error(t *testing.T) {
	var (
		buf         bytes.Buffer
		log, logErr = logger.New(logger.ErrorLevel, logger.JSONFormat, &buf)
	)

	require.NoError(t, logErr)
	assert.Equal(t, logger.ErrorLevel, log.Level())

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	var output = buf.String()

	assert.NotContains(t, output, "debug message")
	assert.NotContains(t, output, "info message")
	assert.NotContains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestLogger_Named_JSONFormat(t *testing.T) {
	var (
		buf    bytes.Buffer
		log, _ = logger.New(logger.DebugLevel, logger.JSONFormat, &buf)
		named  = log.Named("test_name")
	)

	log.Debug("debug message")

	var output = buf.String()

	assert.Contains(t, output, `"msg":"debug message"`)
	assert.NotContains(t, output, `"logger":"`)

	buf.Reset()
	named.Debug("named log message")

	output = buf.String()

	assert.Contains(t, output, `"msg":"named log message"`)
	assert.Contains(t, output, `"logger":"test_name"`)
}

func TestLogger_Named_ConsoleFormat(t *testing.T) {
	var (
		buf    bytes.Buffer
		log, _ = logger.New(logger.DebugLevel, logger.ConsoleFormat, &buf)
		named  = log.Named("test_name")
	)

	log.Debug("debug message")

	var output = buf.String()

	assert.Contains(t, output, `msg="debug message"`)
	assert.NotContains(t, output, `logger=`)

	buf.Reset()
	named.Debug("named log message")

	output = buf.String()

	assert.Contains(t, output, `msg="named log message"`)
	assert.Contains(t, output, `logger=test_name`)
}
