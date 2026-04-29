package logger_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

var (
	rTimestamp = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} `)
	rJSONTs    = regexp.MustCompile(`"ts":[0-9.]+,`)
)

// withoutJSONTs strips the dynamic "ts" field from a JSON log line.
func withoutJSONTs(s string) string { return rJSONTs.ReplaceAllString(s, "") }

func newLog(t *testing.T, buf interface{ Write([]byte) (int, error) }, level logger.Level, format logger.Format) *logger.Logger {
	t.Helper()

	l, err := logger.New(level, format, logger.WithWriter(buf))
	assert.NoError(t, err)

	return l
}

func TestNew_Errors(t *testing.T) {
	t.Parallel()

	_, err := logger.New(logger.Level(127), logger.ConsoleFormat)
	assert.ErrorEqual(t, err, "unsupported logging level")

	_, err = logger.New(logger.WarnLevel, logger.Format(255))
	assert.ErrorEqual(t, err, "unsupported logging format")
}

func TestConsoleFormat(t *testing.T) {
	t.Parallel()

	attrTime := time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)

	t.Run("format", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder

		newLog(t, &buf, logger.DebugLevel, logger.ConsoleFormat).Debug("message",
			logger.String("str", "value"),
			logger.Int("n", 42),
			logger.Bool("ok", true),
			logger.Time("when", attrTime),
			logger.Duration("dur", 500*time.Millisecond),
		)

		want := "DEBUG  message  str=value n=42 ok=true when=2024-01-15T10:30:45.123Z dur=500ms\n"

		assert.Equal(t, want, withoutTimestamps(t, buf.String()))
	})

	t.Run("string quoting", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder

		newLog(t, &buf, logger.DebugLevel, logger.ConsoleFormat).Debug("msg",
			logger.String("plain", "value"),
			logger.String("spaced", "hello world"),
			logger.String("empty", ""),
		)

		want := `DEBUG  msg  plain=value spaced="hello world" empty=""` + "\n"

		assert.Equal(t, want, withoutTimestamps(t, buf.String()))
	})

	t.Run("level filtering", func(t *testing.T) {
		t.Parallel()

		for _, tt := range []struct {
			level logger.Level
			want  string
		}{
			{logger.DebugLevel, "DEBUG  d\nINFO   i\nWARN   w\nERROR  e\n"},
			{logger.InfoLevel, "INFO   i\nWARN   w\nERROR  e\n"},
			{logger.WarnLevel, "WARN   w\nERROR  e\n"},
			{logger.ErrorLevel, "ERROR  e\n"},
		} {
			t.Run(tt.level.String(), func(t *testing.T) {
				t.Parallel()

				var buf strings.Builder

				l := newLog(t, &buf, tt.level, logger.ConsoleFormat)
				l.Debug("d")
				l.Info("i")
				l.Warn("w")
				l.Error("e")

				assert.Equal(t, tt.want, withoutTimestamps(t, buf.String()))
			})
		}
	})

	t.Run("named", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder

		newLog(t, &buf, logger.DebugLevel, logger.ConsoleFormat).Named("svc").Debug("msg")

		assert.Equal(t, "DEBUG  msg  logger=svc\n", withoutTimestamps(t, buf.String()))
	})

	t.Run("with pre-attached attrs", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder

		newLog(t, &buf, logger.DebugLevel, logger.ConsoleFormat).
			With(logger.String("env", "prod")).
			Debug("msg", logger.Int("n", 1))

		assert.Equal(t, "DEBUG  msg  env=prod n=1\n", withoutTimestamps(t, buf.String()))
	})
}

func TestJSONFormat(t *testing.T) {
	t.Parallel()

	attrTime := time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)

	t.Run("format", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder

		newLog(t, &buf, logger.DebugLevel, logger.JSONFormat).Debug("message",
			logger.String("str", "value"),
			logger.Int("n", 42),
			logger.Bool("ok", true),
			logger.Time("when", attrTime),
			logger.Duration("dur", 500*time.Millisecond),
		)

		want := `{"level":"debug","msg":"message","str":"value","n":42,"ok":true,"when":"2024-01-15T10:30:45.123Z","dur":500000000}` + "\n"

		assert.Equal(t, want, withoutJSONTs(buf.String()))
	})

	t.Run("level filtering", func(t *testing.T) {
		t.Parallel()

		for _, tt := range []struct {
			level logger.Level
			want  string
		}{
			{logger.DebugLevel, `{"level":"debug","msg":"d"}` + "\n" + `{"level":"info","msg":"i"}` + "\n" + `{"level":"warn","msg":"w"}` + "\n" + `{"level":"error","msg":"e"}` + "\n"},
			{logger.InfoLevel, `{"level":"info","msg":"i"}` + "\n" + `{"level":"warn","msg":"w"}` + "\n" + `{"level":"error","msg":"e"}` + "\n"},
			{logger.WarnLevel, `{"level":"warn","msg":"w"}` + "\n" + `{"level":"error","msg":"e"}` + "\n"},
			{logger.ErrorLevel, `{"level":"error","msg":"e"}` + "\n"},
		} {
			t.Run(tt.level.String(), func(t *testing.T) {
				t.Parallel()

				var buf strings.Builder

				l := newLog(t, &buf, tt.level, logger.JSONFormat)
				l.Debug("d")
				l.Info("i")
				l.Warn("w")
				l.Error("e")

				assert.Equal(t, tt.want, withoutJSONTs(buf.String()))
			})
		}
	})

	t.Run("named", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder

		newLog(t, &buf, logger.DebugLevel, logger.JSONFormat).Named("svc").Debug("msg")

		assert.Contains(t, buf.String(), `"logger":"svc"`)
	})
}

func TestNewNop(t *testing.T) {
	t.Parallel()

	l := logger.NewNop()
	assert.NotNil(t, l)
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	l.Error("e")
	_ = l.Named("x")
	_ = l.With(logger.String("k", "v"))
}

// withoutTimestamps validates and strips the "HH:MM:SS.mmm " prefix from every log line in s.
func withoutTimestamps(t *testing.T, s string) string {
	t.Helper()

	const n = len("00:00:00.000 ")

	var b strings.Builder

	for _, line := range strings.SplitAfter(s, "\n") {
		if line == "" {
			continue
		}

		if !rTimestamp.MatchString(line) {
			t.Fatalf("missing HH:MM:SS.mmm timestamp prefix: %q", line)
		}

		b.WriteString(line[n:])
	}

	return b.String()
}
