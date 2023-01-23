package logger

import (
	"fmt"
	"strings"
)

// A Format is a logging format.
type Format uint8

const (
	ConsoleFormat Format = iota // useful for console output (for humans)
	JSONFormat                  // useful for logging aggregation systems (for robots)
)

// String returns a lower-case ASCII representation of the log format.
func (f Format) String() string {
	switch f {
	case ConsoleFormat:
		return "console"
	case JSONFormat:
		return "json"
	}

	return fmt.Sprintf("format(%d)", f)
}

// Formats returns a slice of all logging formats.
func Formats() []Format {
	return []Format{ConsoleFormat, JSONFormat}
}

// FormatStrings returns a slice of all logging formats as strings.
func FormatStrings() []string {
	var (
		formats = Formats()
		result  = make([]string, len(formats))
	)

	for i := range formats {
		result[i] = formats[i].String()
	}

	return result
}

// ParseFormat parses a format (case is ignored) based on the ASCII representation of the log format.
// If the provided ASCII representation is invalid an error is returned.
//
// This is particularly useful when dealing with text input to configure log formats.
func ParseFormat[T string | []byte](text T) (Format, error) {
	var format string

	if s, ok := any(text).(string); ok {
		format = s
	} else {
		format = string(any(text).([]byte))
	}

	switch strings.ToLower(format) {
	case "console", "": // make the zero value useful
		return ConsoleFormat, nil
	case "json":
		return JSONFormat, nil
	}

	return Format(0), fmt.Errorf("unrecognized logging format: %q", text)
}
