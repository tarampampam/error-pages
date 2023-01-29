package logger

import (
	"fmt"
	"strings"
)

// A Level is a logging level.
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel        // default level (zero-value)
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns a lower-case ASCII representation of the log level.
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
	case FatalLevel:
		return "fatal"
	}

	return fmt.Sprintf("level(%d)", l)
}

// Levels returns a slice of all logging levels.
func Levels() []Level {
	return []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
}

// LevelStrings returns a slice of all logging levels as strings.
func LevelStrings() []string {
	var (
		levels = Levels()
		result = make([]string, len(levels))
	)

	for i := range levels {
		result[i] = levels[i].String()
	}

	return result
}

// ParseLevel parses a level (case is ignored) based on the ASCII representation of the log level.
// If the provided ASCII representation is invalid an error is returned.
//
// This is particularly useful when dealing with text input to configure log levels.
func ParseLevel[T string | []byte](text T) (Level, error) {
	var lvl string

	if s, ok := any(text).(string); ok {
		lvl = s
	} else {
		lvl = string(any(text).([]byte))
	}

	switch strings.ToLower(lvl) {
	case "debug", "verbose", "trace":
		return DebugLevel, nil
	case "info", "": // make the zero value useful
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	}

	return Level(0), fmt.Errorf("unrecognized logging level: %q", text)
}
