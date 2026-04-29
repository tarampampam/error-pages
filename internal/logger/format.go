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

// String returns the format name in lowercase (e.g. "console", "json").
func (f Format) String() string {
	switch f {
	case ConsoleFormat:
		return "console"
	case JSONFormat:
		return "json"
	}

	return fmt.Sprintf("format(%d)", f)
}

// ParseFormat converts a string name to a Format (case-insensitive).
// Returns an error if the name is not recognized.
func ParseFormat(text string) (Format, error) {
	switch strings.ToLower(text) {
	case "console", "": // make the zero value useful
		return ConsoleFormat, nil
	case "json":
		return JSONFormat, nil
	}

	return Format(0), fmt.Errorf("unrecognized logging format: %q", text)
}
