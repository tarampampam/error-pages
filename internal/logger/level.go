package logger

import (
	"fmt"
	"log/slog"
	"strings"
)

// Level is a logging priority. Higher values are more severe.
type Level int

const (
	DebugLevel = Level(slog.LevelDebug) // -4
	InfoLevel  = Level(slog.LevelInfo)  // 0; the zero value of Level
	WarnLevel  = Level(slog.LevelWarn)  // 4
	ErrorLevel = Level(slog.LevelError) // 8
)

// String returns the level name in lowercase (e.g. "debug", "info").
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	}

	return fmt.Sprintf("level(%d)", l)
}

// ParseLevel converts a string name to a Level (case-insensitive).
// Returns an error if the name is not recognized.
func ParseLevel(text string) (Level, error) {
	switch strings.ToLower(text) {
	case "debug", "verbose", "trace": // verbose and trace are treated as aliases for debug
		return DebugLevel, nil
	case "info", "": // make the zero value useful
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error", "err":
		return ErrorLevel, nil
	}

	return Level(0), fmt.Errorf("unrecognized logging level: %q", text)
}
