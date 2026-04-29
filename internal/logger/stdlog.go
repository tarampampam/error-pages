package logger

import (
	"log"
	"strings"
)

// NewStdLog returns a [log.Logger] that forwards all writes to the provided logger at the given level.
//
// The primary use case is wiring third-party components that accept a [log.Logger] (e.g. an HTTP
// server's ErrorLog field) into the application's structured logger:
//
//	srv := &http.Server{
//	    ErrorLog: logger.NewStdLog(log.Named("http.server"), logger.WarnLevel),
//	}
//
// Each write from the returned logger is emitted as a single structured record. The prefix and
// flags of the underlying [log.Logger] are intentionally left empty to avoid duplicating
// timestamp/level information that slog already provides.
func NewStdLog(logger *Logger, level Level) *log.Logger {
	// Empty prefix: log.Logger prepends the prefix to every message before calling Write, so any non-empty
	// prefix would be embedded literally in the slog "msg" field instead of appearing as a structured attribute.
	//
	// Zero flags: log.LstdFlags (the default) prepends a formatted date+time to every message before Write
	// is called, which would duplicate the timestamp that slog already emits.
	return log.New(&loggerWriter{log: logger, level: level}, "", 0)
}

type loggerWriter struct {
	log   *Logger
	level Level
}

func (lw *loggerWriter) Write(p []byte) (int, error) {
	// log.Logger always appends \n; strip it to avoid a trailing newline in the slog message.
	lw.log.Log(lw.level, strings.TrimSuffix(string(p), "\n"))

	return len(p), nil
}
