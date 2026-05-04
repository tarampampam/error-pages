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

// Logger is a structured logger with a zap-like API, but with [slog.Logger] under the hood.
type Logger struct {
	log  *slog.Logger
	lvl  Level
	name string // set by [Logger.Named] or [WithName]; stored here instead of in the handler to allow chaining
}

type (
	options struct {
		Writer io.Writer
		Name   string
	}

	// Option is a functional option for configuring the logger.
	Option func(*options)
)

// WithWriter sets the writer for log output. Defaults to [os.Stderr].
func WithWriter(w io.Writer) Option { return func(o *options) { o.Writer = w } }

// WithName sets the initial logger name. Equivalent to calling [Logger.Named] on the returned logger.
func WithName(name string) Option { return func(o *options) { o.Name = name } }

// New creates a new logger with the given level and format.
// Use [WithWriter] to override the default output ([os.Stderr]) and [WithName] to set an initial name.
func New(l Level, f Format, opt ...Option) (*Logger, error) {
	opts := options{Writer: os.Stderr}

	for _, o := range opt {
		o(&opts)
	}

	var handlerOptions slog.HandlerOptions

	switch l {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel:
		handlerOptions.Level = slog.Level(l)
	default:
		return nil, errors.New("unsupported logging level")
	}

	var handler slog.Handler

	switch f {
	case ConsoleFormat:
		handler = newConsoleHandler(opts.Writer, slog.Level(l))
	case JSONFormat:
		handlerOptions.ReplaceAttr = replaceAttrJSON
		handler = slog.NewJSONHandler(opts.Writer, &handlerOptions)
	default:
		return nil, errors.New("unsupported logging format")
	}

	return &Logger{log: slog.New(handler), lvl: l, name: opts.Name}, nil
}

// Level returns the minimum level at which the logger emits records.
func (l *Logger) Level() Level { return l.lvl }

// Enabled reports whether the logger would log a record at the given level.
func (l *Logger) Enabled(level Level) bool { return level >= l.lvl }

// Named creates a new logger with the given name attached as a "logger" field on every record.
// When called on an already-named logger, names are joined with a dot: Named("a").Named("b") emits logger=a.b.
func (l *Logger) Named(name string) *Logger {
	if name == "" {
		return l
	}

	joined := name
	if l.name != "" {
		joined = l.name + "." + name
	}

	return &Logger{log: l.log, lvl: l.lvl, name: joined}
}

// With creates a new logger with the given fields attached to every subsequent record.
func (l *Logger) With(f ...Attr) *Logger {
	if len(f) == 0 {
		return l
	}

	return &Logger{log: slog.New(l.log.Handler().WithAttrs(f)), lvl: l.lvl, name: l.name}
}

// Log logs a message at the given level. Use it directly when the level is determined at runtime.
//
//nolint:contextcheck,nolintlint // context is intentionally not threaded through log calls
func (l *Logger) Log(level Level, msg string, f ...Attr) {
	slogLevel := slog.Level(level)

	ctx := context.Background()

	// slog accepts a context so that custom handlers can extract request-scoped values (e.g. trace IDs)
	// from it. We use Background() to keep the API simple - no context threading required at call sites
	if l.name == "" {
		l.log.LogAttrs(ctx, slogLevel, msg, filterAttrs(f)...)

		return
	}

	// Guard the allocation: skip if the level is disabled.
	if !l.log.Enabled(ctx, slogLevel) {
		return
	}

	filtered := filterAttrs(f)
	all := make([]Attr, 0, len(filtered)+1)
	all = append(all, loggerNameAttr(l.name))
	all = append(all, filtered...)

	l.log.LogAttrs(ctx, slogLevel, msg, all...)
}

// Debug logs a message at DebugLevel.
func (l *Logger) Debug(msg string, f ...Attr) { l.Log(DebugLevel, msg, f...) }

// Info logs a message at InfoLevel.
func (l *Logger) Info(msg string, f ...Attr) { l.Log(InfoLevel, msg, f...) }

// Warn logs a message at WarnLevel.
func (l *Logger) Warn(msg string, f ...Attr) { l.Log(WarnLevel, msg, f...) }

// Error logs a message at ErrorLevel.
func (l *Logger) Error(msg string, f ...Attr) { l.Log(ErrorLevel, msg, f...) }

// loggerNameAttr returns the "logger" attribute for the given name.
func loggerNameAttr(loggerName string) Attr { return slog.String("logger", loggerName) }

// filterAttrs returns f with empty Attrs removed. Uses a copy-on-first-match pattern: the input slice
// is returned as-is (no allocation) when no empty Attrs are present; a new slice is allocated only on
// the first empty Attr found.
func filterAttrs(f []Attr) []Attr {
	for firstEmpty, attr := range f {
		if !attr.Equal(slog.Attr{}) {
			continue
		}

		// all attrs before firstEmpty are non-empty (the loop would have triggered earlier otherwise),
		// so copy them directly into the output slice without re-checking
		out := make([]Attr, firstEmpty, len(f))
		copy(out, f[:firstEmpty])

		// scan the rest, keeping only non-empty attrs
		for _, remaining := range f[firstEmpty+1:] {
			if !remaining.Equal(slog.Attr{}) {
				out = append(out, remaining)
			}
		}

		return out
	}

	return f
}

// lowerCaseLevel converts the built-in slog level attribute value to lowercase.
func lowerCaseLevel(a slog.Attr) slog.Attr {
	return slog.String(a.Key, strings.ToLower(a.Value.String()))
}

// replaceAttrJSON is the ReplaceAttr function for the JSON format.
func replaceAttrJSON(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.LevelKey:
		return lowerCaseLevel(a)
	case slog.TimeKey:
		if ts, ok := a.Value.Any().(time.Time); ok {
			return slog.Float64("ts", float64(ts.Unix())+float64(ts.Nanosecond())/1e9)
		}
	}

	return a
}
