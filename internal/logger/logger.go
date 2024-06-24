package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// internalAttrKeyLoggerName is used to store the logger name in the logger context (attributes).
const internalAttrKeyLoggerName = "named_logger"

var (
	// consoleFormatAttrReplacer is a replacer for console format. It replaces some attributes with more
	// human-readable ones.
	consoleFormatAttrReplacer = func(_ []string, a slog.Attr) slog.Attr { //nolint:gochecknoglobals
		switch a.Key {
		case internalAttrKeyLoggerName:
			return slog.String("logger", a.Value.String())
		case "level":
			return slog.String(a.Key, strings.ToLower(a.Value.String()))
		default:
			if ts, ok := a.Value.Any().(time.Time); ok && a.Key == "time" {
				return slog.String(a.Key, ts.Format("15:04:05"))
			}
		}

		return a
	}

	// jsonFormatAttrReplacer is a replacer for JSON format. It replaces some attributes with more
	// machine-readable ones.
	jsonFormatAttrReplacer = func(_ []string, a slog.Attr) slog.Attr { //nolint:gochecknoglobals
		switch a.Key {
		case internalAttrKeyLoggerName:
			return slog.String("logger", a.Value.String())
		case "level":
			return slog.String(a.Key, strings.ToLower(a.Value.String()))
		default:
			if ts, ok := a.Value.Any().(time.Time); ok && a.Key == "time" {
				return slog.Float64("ts", float64(ts.Unix())+float64(ts.Nanosecond())/1e9)
			}
		}

		return a
	}
)

// Logger is a simple logger that wraps [slog.Logger]. It provides a more convenient API for logging and
// formatting messages.
type Logger struct {
	ctx  context.Context
	slog *slog.Logger
	lvl  Level
}

// New creates a new logger with the given level and format. Optionally, you can specify the writer to write logs to.
func New(l Level, f Format, writer ...io.Writer) (*Logger, error) {
	var options slog.HandlerOptions

	switch l {
	case DebugLevel:
		options.Level = slog.LevelDebug
	case InfoLevel:
		options.Level = slog.LevelInfo
	case WarnLevel:
		options.Level = slog.LevelWarn
	case ErrorLevel:
		options.Level = slog.LevelError
	default:
		return nil, errors.New("unsupported logging level")
	}

	var (
		handler slog.Handler
		target  io.Writer
	)

	if len(writer) > 0 && writer[0] != nil {
		target = writer[0]
	} else {
		target = os.Stderr
	}

	switch f {
	case ConsoleFormat:
		options.ReplaceAttr = consoleFormatAttrReplacer

		handler = slog.NewTextHandler(target, &options)
	case JSONFormat:
		options.ReplaceAttr = jsonFormatAttrReplacer

		handler = slog.NewJSONHandler(target, &options)
	default:
		return nil, errors.New("unsupported logging format")
	}

	return &Logger{ctx: context.Background(), slog: slog.New(handler), lvl: l}, nil
}

// Level returns the logger level.
func (l *Logger) Level() Level { return l.lvl }

// Named creates a new logger with the same properties as the original logger and the given name.
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		ctx:  l.ctx,
		slog: l.slog.With(slog.String(internalAttrKeyLoggerName, name)),
		lvl:  l.lvl,
	}
}

// Debug logs a message at DebugLevel.
func (l *Logger) Debug(msg string, f ...Attr) { l.slog.LogAttrs(l.ctx, slog.LevelDebug, msg, f...) }

// Info logs a message at InfoLevel.
func (l *Logger) Info(msg string, f ...Attr) { l.slog.LogAttrs(l.ctx, slog.LevelInfo, msg, f...) }

// Warn logs a message at WarnLevel.
func (l *Logger) Warn(msg string, f ...Attr) { l.slog.LogAttrs(l.ctx, slog.LevelWarn, msg, f...) }

// Error logs a message at ErrorLevel.
func (l *Logger) Error(msg string, f ...Attr) { l.slog.LogAttrs(l.ctx, slog.LevelError, msg, f...) }
