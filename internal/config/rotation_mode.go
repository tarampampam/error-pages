package config

import (
	"fmt"
	"strings"
)

// RotationMode represents the rotation mode for templates.
type RotationMode byte

const (
	RotationModeDisabled            RotationMode = iota // do not rotate templates, default
	RotationModeRandomOnStartup                         // pick a random template on startup
	RotationModeRandomOnEachRequest                     // pick a random template on each request
	RotationModeRandomDaily                             // once a day switch to a random template
	RotationModeRandomHourly                            // once an hour switch to a random template
)

// String returns a human-readable representation of the rotation mode.
func (rm RotationMode) String() string {
	switch rm {
	case RotationModeDisabled:
		return "disabled"
	case RotationModeRandomOnStartup:
		return "random-on-startup"
	case RotationModeRandomOnEachRequest:
		return "random-on-each-request"
	case RotationModeRandomDaily:
		return "random-daily"
	case RotationModeRandomHourly:
		return "random-hourly"
	}

	return fmt.Sprintf("RotationMode(%d)", rm)
}

// RotationModes returns a slice of all rotation modes.
func RotationModes() []RotationMode {
	return []RotationMode{
		RotationModeDisabled,
		RotationModeRandomOnStartup,
		RotationModeRandomOnEachRequest,
		RotationModeRandomDaily,
		RotationModeRandomHourly,
	}
}

// RotationModeStrings returns a slice of all rotation modes as strings.
func RotationModeStrings() []string {
	var (
		modes  = RotationModes()
		result = make([]string, len(modes))
	)

	for i := range modes {
		result[i] = modes[i].String()
	}

	return result
}

// ParseRotationMode parses a rotation mode (case is ignored) based on the ASCII representation of the rotation mode.
// If the provided ASCII representation is invalid an error is returned.
func ParseRotationMode[T string | []byte](text T) (RotationMode, error) {
	var mode string

	if s, ok := any(text).(string); ok {
		mode = s
	} else {
		mode = string(any(text).([]byte))
	}

	switch strings.ToLower(mode) {
	case RotationModeDisabled.String(), "":
		return RotationModeDisabled, nil // the empty string makes sense
	case RotationModeRandomOnStartup.String():
		return RotationModeRandomOnStartup, nil
	case RotationModeRandomOnEachRequest.String():
		return RotationModeRandomOnEachRequest, nil
	case RotationModeRandomDaily.String():
		return RotationModeRandomDaily, nil
	case RotationModeRandomHourly.String():
		return RotationModeRandomHourly, nil
	}

	return RotationMode(0), fmt.Errorf("unrecognized rotation mode: %q", mode)
}
