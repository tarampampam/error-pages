package logger

import "log/slog"

// NewNop returns a no-op Logger that discards all records without any overhead.
// The common use case is to use it in tests.
func NewNop() *Logger {
	return &Logger{log: slog.New(slog.DiscardHandler), lvl: DebugLevel}
}
