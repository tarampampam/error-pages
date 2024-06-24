package logger

import (
	stdLog "log"
)

// NewStdLog returns a *[log.Logger] which writes to the supplied [Logger] at [InfoLevel].
func NewStdLog(log *Logger) *stdLog.Logger {
	return stdLog.New(&loggerWriter{log} /* prefix */, "" /* flags */, 0)
}

type loggerWriter struct{ log *Logger }

func (lw *loggerWriter) Write(p []byte) (int, error) { lw.log.Info(string(p)); return len(p), nil } //nolint:nlreturn
