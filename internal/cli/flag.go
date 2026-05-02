package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type (
	// Flagger defines an interface for command-line flags with methods for
	// checking if the flag was set, retrieving help text, applying the flag to
	// a flag set, validating the value, and executing an associated action.
	Flagger interface {
		IsSet() bool                        // Checks if the flag was explicitly set.
		Help() (names string, usage string) // Returns the flag's names and usage description.
		Apply(*flag.FlagSet)                // Registers the flag with a flag set.
		Validate(*Command) error            // Validates the flag's value.
		RunAction(*Command) error           // Executes an associated action if set.
	}

	// FlagType defines supported data types for flags.
	FlagType interface {
		bool | int | int64 | string | uint | uint64 | float64 | time.Duration
	}

	// Flag represents a command-line flag with metadata and behavior.
	Flag[T FlagType] struct {
		Names        []string                // Flag names (e.g., ["config-file", "c"]).
		Usage        string                  // Flag description (e.g., "Path to the configuration file").
		Default      T                       // Default value of the flag.
		EnvVars      []string                // Environment variable names for this flag (e.g., ["CONFIG_FILE"]).
		Validator    func(*Command, T) error // Optional function to validate the value.
		Action       func(*Command, T) error // Optional function to execute when the flag is set.
		ValueSetFrom flagValueSource         // Source of the value (default, env, CLI flag).
		Value        *T                      // Pointer to store the parsed flag value.
	}
)

// Ensures that Flag[T] implements the Flagger interface for all supported types.
var (
	_ Flagger = (*Flag[bool])(nil)
	_ Flagger = (*Flag[int])(nil)
	_ Flagger = (*Flag[int64])(nil)
	_ Flagger = (*Flag[string])(nil)
	_ Flagger = (*Flag[uint])(nil)
	_ Flagger = (*Flag[uint64])(nil)
	_ Flagger = (*Flag[float64])(nil)
	_ Flagger = (*Flag[time.Duration])(nil)
)

type flagValueSource = byte

// Enumerates possible sources for a flag's value.
const (
	FlagValueSourceNone    flagValueSource = iota // Value not set.
	FlagValueSourceDefault                        // Value set from default.
	FlagValueSourceEnv                            // Value set from environment variable.
	FlagValueSourceFlag                           // Value set from command-line flag.
)

// IsSet checks if the flag was explicitly set (i.e., not using the default value).
func (f *Flag[T]) IsSet() bool {
	if f.Value == nil {
		return false // flag was never assigned a value
	}

	switch f.ValueSetFrom {
	case FlagValueSourceNone, FlagValueSourceDefault:
		return false
	default:
		return *f.Value != f.Default // true if value differs from default
	}
}

// Help returns a formatted flag name string and usage description.
func (f *Flag[T]) Help() (names string, usage string) {
	var b strings.Builder

	b.Grow(len(f.Usage))

	// construct the flag names with proper prefixes
	for i, name := range f.Names {
		if i > 0 {
			b.WriteString(", ")
		}

		if len(name) == 1 {
			b.WriteRune('-') // single-character flags use a single dash
		} else {
			b.WriteString("--") // long flags use double dashes
		}

		b.WriteString(name)

		// boolean flags don't require an explicit value
		if _, ok := any(*new(T)).(bool); !ok {
			b.WriteString(`="…"`)
		}
	}

	names = b.String()

	b.Reset()
	b.WriteString(f.Usage)

	// append default value if present
	if f.Default != *new(T) {
		if b.Len() > 0 {
			b.WriteRune(' ')
		}

		b.WriteString("(default: ")
		_, _ = fmt.Fprintf(&b, "%v", f.Default)
		b.WriteRune(')')
	}

	// append environment variable names if present
	if len(f.EnvVars) > 0 {
		if b.Len() > 0 {
			b.WriteRune(' ')
		}

		b.WriteRune('[')

		for i, envVar := range f.EnvVars {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteRune('$')
			b.WriteString(envVar)
		}

		b.WriteRune(']')
	}

	return names, b.String()
}

// predefined errors for invalid flag values.
var (
	errInvalidBool     = errors.New("must be a valid boolean value (e.g., true/false, 1/0)")
	errInvalidInt      = errors.New("must contain only digits with an optional leading '-'")
	errInvalidUint     = errors.New("must contain only digits (positive numbers only)")
	errInvalidFloat    = errors.New("must contain only digits with an optional decimal point")
	errInvalidDuration = errors.New("must be a valid Go duration string (e.g., 1h30m, -2s, 500ms)")
)

// parseString converts a string to the corresponding flag type.
func (f *Flag[T]) parseString(s string) (T, error) {
	var empty T // default zero value of type T

	// cast converts a concrete value to T; each branch is guarded by the type switch above,
	// so the assertion will never fail in practice.
	cast := func(v any) (T, error) {
		result, ok := v.(T)
		if !ok {
			return empty, fmt.Errorf("unreachable: type assertion to %T failed", empty)
		}

		return result, nil
	}

	switch any(empty).(type) {
	case bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return empty, errInvalidBool
		}

		return cast(v)
	case int:
		v, err := strconv.Atoi(s)
		if err != nil {
			return empty, errInvalidInt
		}

		return cast(v)
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return empty, errInvalidInt
		}

		return cast(v)
	case string:
		return cast(s)
	case uint:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return empty, errInvalidUint
		}

		return cast(uint(v))
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return empty, errInvalidUint
		}

		return cast(v)
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return empty, errInvalidFloat
		}

		return cast(v)
	case time.Duration:
		v, err := time.ParseDuration(s)
		if err != nil {
			return empty, errInvalidDuration
		}

		return cast(v)
	}

	return empty, fmt.Errorf("unsupported flag type: %T", empty) // will never happen
}

// envValue retrieves the flag value from environment variables, if set.
func (f *Flag[T]) envValue() (
	value T, // the value
	found bool, // a boolean flag indicating whether the value was found
	envName string, // the name of the environment variable
	_ error, // an error if any
) {
	var empty T // default zero value

	for _, name := range f.EnvVars {
		if envValue, ok := os.LookupEnv(name); ok {
			envValue = strings.Trim(envValue, " \t\n\r") // remove surrounding whitespace

			parsed, err := f.parseString(envValue)
			if err != nil {
				return empty, true, name, err // return error if parsing fails
			}

			return parsed, true, name, nil // successfully parsed value
		}
	}

	return empty, false, "", nil // no environment variable found
}

// setValue assigns the flag's value and records its source.
func (f *Flag[T]) setValue(v T, src flagValueSource) {
	if f.Value == nil {
		f.Value = new(T)
	}

	*f.Value, f.ValueSetFrom = v, src
}

// Apply registers the flag with the provided flag set.
func (f *Flag[T]) Apply(s *flag.FlagSet) {
	// set the default flag value
	f.setValue(f.Default, FlagValueSourceDefault)

	// attempt to load from environment variable (note: parsing errors are ignored)
	if v, found, _, err := f.envValue(); found && err == nil {
		f.setValue(v, FlagValueSourceEnv)
	}

	switch any(*new(T)).(type) {
	case bool:
		var fn = func(string) error {
			// since we have a boolean flag, we need to set the value to true if the flag was provided
			// without taking into account the value
			if result, ok := any(true).(T); ok {
				f.setValue(result, FlagValueSourceFlag)
			}

			return nil
		}

		for _, name := range f.Names {
			s.BoolFunc(name, f.Usage, fn)
		}
	default:
		var fn = func(in string) error {
			if v, parsingErr := f.parseString(in); parsingErr == nil {
				f.setValue(v, FlagValueSourceFlag)
			} else {
				return parsingErr
			}

			return nil
		}

		for _, name := range f.Names {
			s.Func(name, f.Usage, fn)
		}
	}
}

// Validate checks if the flag's value is valid.
func (f *Flag[T]) Validate(c *Command) error {
	if f.Validator == nil {
		return nil
	}

	if f.Value == nil {
		return errors.New("flag value is nil")
	}

	return f.Validator(c, *f.Value)
}

// RunAction runs the flag's action if set.
func (f *Flag[T]) RunAction(c *Command) error {
	if f.Action == nil {
		return nil
	}

	if f.Value == nil {
		return errors.New("flag value is nil")
	}

	return f.Action(c, *f.Value)
}
