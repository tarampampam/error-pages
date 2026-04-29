package logger

import (
	"context"
	"log/slog"
	"sync"
)

// Record is a captured log record produced by a [Recorder]-backed Logger.
type Record struct {
	Level   Level
	Message string
	Attrs   map[string]slog.Value
}

// Recorder captures log records emitted by a Logger for use in tests.
type Recorder struct {
	mu      sync.Mutex
	records []Record
}

// Records returns a snapshot of all captured records in the order they were received.
func (r *Recorder) Records() []Record {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]Record(nil), r.records...)
}

// Len returns the number of captured records.
func (r *Recorder) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.records)
}

// NewRecorder creates a Logger backed by an in-memory handler and returns both the logger and the recorder.
// The logger captures records at any level. Intended for use in tests.
func NewRecorder() (*Logger, *Recorder) {
	rec := &Recorder{}

	return &Logger{log: slog.New(&recorderHandler{rec: rec}), lvl: DebugLevel}, rec
}

// recorderHandler is a [slog.Handler] that appends records to a Recorder.
type recorderHandler struct {
	rec      *Recorder
	preAttrs []slog.Attr
}

var _ slog.Handler = (*recorderHandler)(nil)

func (h *recorderHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *recorderHandler) Handle(_ context.Context, rec slog.Record) error {
	attrs := make(map[string]slog.Value, rec.NumAttrs()+len(h.preAttrs))

	for _, a := range h.preAttrs {
		attrs[a.Key] = a.Value
	}

	rec.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value

		return true
	})

	h.rec.mu.Lock()
	h.rec.records = append(h.rec.records, Record{
		Level:   Level(rec.Level),
		Message: rec.Message,
		Attrs:   attrs,
	})
	h.rec.mu.Unlock()

	return nil
}

func (h *recorderHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := make([]slog.Attr, len(h.preAttrs)+len(attrs))
	copy(combined, h.preAttrs)
	copy(combined[len(h.preAttrs):], attrs)

	return &recorderHandler{rec: h.rec, preAttrs: combined}
}

func (h *recorderHandler) WithGroup(_ string) slog.Handler { return h }
