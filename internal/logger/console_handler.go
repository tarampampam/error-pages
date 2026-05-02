package logger

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strconv"
	"sync"
)

var _ slog.Handler = (*consoleHandler)(nil)

// consoleHandler is a [slog.Handler] that writes a single human-readable line per record.
//
// Line format:
//
//	HH:MM:SS.mmm LEVEL  message  key=value
//
// The level field is padded to five characters so messages stay aligned across levels.
// String attribute values are double-quoted only when they contain whitespace, a backslash,
// or a double-quote. All other value types use their natural string representation.
type consoleHandler struct {
	mu       sync.Mutex
	w        io.Writer
	level    slog.Level
	preAttrs []byte   // pre-formatted " key=value …" chunks from [consoleHandler.WithAttrs]
	groups   []string // active group stack from [consoleHandler.WithGroup]
}

func newConsoleHandler(w io.Writer, level slog.Level) *consoleHandler {
	return &consoleHandler{w: w, level: level}
}

func (h *consoleHandler) clone() *consoleHandler {
	return &consoleHandler{
		w:        h.w,
		level:    h.level,
		preAttrs: bytes.Clone(h.preAttrs),
		groups:   append([]string(nil), h.groups...),
	}
}

func (h *consoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// WithAttrs returns a new handler with attrs pre-formatted using the active group stack.
func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	clone := h.clone()

	var buf bytes.Buffer
	buf.Write(clone.preAttrs)

	for _, a := range attrs {
		appendConsoleAttr(&buf, clone.groups, a)
	}

	clone.preAttrs = buf.Bytes()

	return clone
}

// WithGroup returns a new handler with name appended to the group stack.
// Subsequent [consoleHandler.WithAttrs] calls will prefix their keys with the accumulated group names.
func (h *consoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	clone := h.clone()
	clone.groups = append(clone.groups, name)

	return clone
}

// Handle writes one log line to the writer.
func (h *consoleHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	// HH:MM:SS.mmm - omitted when the record carries no timestamp.
	if !r.Time.IsZero() {
		buf.WriteString(r.Time.Format("15:04:05.000"))
		buf.WriteByte(' ')
	}

	// Level padded to 5 chars: DEBUG=5, ERROR=5, INFO/WARN=4 → padded to 5.
	level := r.Level.String()
	buf.WriteString(level)

	for i := len(level); i < 5; i++ {
		buf.WriteByte(' ')
	}

	buf.WriteString("  ")

	// Message.
	buf.WriteString(r.Message)

	// Collect all attrs so the separator is only emitted when there is something to show.
	var attrBuf bytes.Buffer

	attrBuf.Write(h.preAttrs)
	r.Attrs(func(a slog.Attr) bool {
		appendConsoleAttr(&attrBuf, h.groups, a)

		return true
	})

	if attrBuf.Len() > 0 {
		buf.WriteByte(' ') // attrBuf already starts with ' ', so together = "  "
		buf.Write(attrBuf.Bytes())
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.w.Write(buf.Bytes())

	return err
}

// appendConsoleAttr writes " key=value" (with a leading space) to buf.
// Group-valued attrs are expanded: each child is written with the group name prepended to its key.
func appendConsoleAttr(buf *bytes.Buffer, groups []string, a slog.Attr) {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return
	}

	if a.Value.Kind() == slog.KindGroup {
		subAttrs := a.Value.Group()
		if len(subAttrs) == 0 {
			return
		}

		// An empty key means the group's children are inlined without a prefix (slog spec).
		// Otherwise cap-trick the slice so the append always allocates a fresh backing array
		// and recursive calls cannot stomp each other's group stack.
		sub := groups
		if a.Key != "" {
			sub = append(groups[:len(groups):len(groups)], a.Key)
		}

		for _, sa := range subAttrs {
			appendConsoleAttr(buf, sub, sa)
		}

		return
	}

	buf.WriteByte(' ')

	for _, g := range groups {
		buf.WriteString(g)
		buf.WriteByte('.')
	}

	buf.WriteString(a.Key)
	buf.WriteByte('=')
	appendConsoleValue(buf, a.Value)
}

// appendConsoleValue formats v into buf.
// String values are quoted only when necessary; time values use millisecond precision;
// everything else uses the slog default string representation.
func appendConsoleValue(buf *bytes.Buffer, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		s := v.String()
		if needsConsoleQuoting(s) {
			buf.WriteString(strconv.Quote(s))
		} else {
			buf.WriteString(s)
		}
	case slog.KindTime:
		buf.WriteString(v.Time().Format("2006-01-02T15:04:05.000Z07:00"))
	case slog.KindAny, slog.KindBool, slog.KindDuration, slog.KindFloat64, slog.KindInt64, slog.KindUint64,
		slog.KindGroup, slog.KindLogValuer:
		buf.WriteString(v.String())
	}
}

// needsConsoleQuoting reports whether s must be wrapped in double-quotes for readable output.
func needsConsoleQuoting(s string) bool {
	if s == "" {
		return true
	}

	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '"' || c == '\\' {
			return true
		}
	}

	return false
}
