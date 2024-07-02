package logger

import (
	"context"
	"log/slog"
)

// NewNop returns a no-op Logger. It never writes out logs or internal errors. The common use case is to use it
// in tests.
func NewNop() *Logger {
	return &Logger{ctx: context.Background(), slog: slog.New(&noopHandler{}), lvl: DebugLevel}
}

type noopHandler struct{}

var _ slog.Handler = (*noopHandler)(nil) // verify interface implementation

func (noopHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (noopHandler) Handle(context.Context, slog.Record) error { return nil }
func (noopHandler) WithAttrs([]slog.Attr) slog.Handler        { return noopHandler{} }
func (noopHandler) WithGroup(string) slog.Handler             { return noopHandler{} }
