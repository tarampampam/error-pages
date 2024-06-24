package logger

import (
	"log/slog"
	"time"
)

// An Attr is a key-value pair.
type Attr = slog.Attr

// String returns an Attr for a string value.
func String(key, value string) Attr { return slog.String(key, value) }

// Strings returns an Attr for a slice of strings.
func Strings(key string, value ...string) Attr { return slog.Any(key, value) }

// Int64 returns an Attr for an int64.
func Int64(key string, value int64) Attr { return slog.Int64(key, value) }

// Int converts an int to an int64 and returns an Attr with that value.
func Int(key string, value int) Attr { return slog.Int(key, value) }

// Uint64 returns an Attr for an uint64.
func Uint64(key string, v uint64) Attr { return slog.Uint64(key, v) }

// Uint16 returns an Attr for an uint16.
func Uint16(key string, v uint16) Attr { return slog.Uint64(key, uint64(v)) }

// Float64 returns an Attr for a floating-point number.
func Float64(key string, v float64) Attr { return slog.Float64(key, v) }

// Bool returns an Attr for a bool.
func Bool(key string, v bool) Attr { return slog.Bool(key, v) }

// Time returns an Attr for a [time.Time]. It discards the monotonic portion.
func Time(key string, v time.Time) Attr { return slog.Time(key, v) }

// Duration returns an Attr for a [time.Duration].
func Duration(key string, v time.Duration) Attr { return slog.Duration(key, v) }

// Any returns an Attr for any value.
func Any(key string, v any) Attr { return slog.Any(key, v) }
