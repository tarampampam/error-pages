package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"
)

// Command represents a CLI command with flags, description, usage, and an action function.
type Command struct {
	Name        string    // Name of the command.
	Description string    // Brief description of the command.
	Usage       string    // Usage example of the command.
	Version     string    // Version of the command.
	Flags       []Flagger // Collection of flags associated with the command.
	Output      io.Writer // Output writer, defaults to os.Stdout if not set.

	Action func(_ context.Context, _ *Command, args []string) error // Action function executed when the command runs.

	initOnce              sync.Once // to ensure initialization is done only once
	showHelp, showVersion bool      // built-in flags for displaying help and version
}

func (c *Command) init() {
	c.initOnce.Do(func() {
		c.Flags = append(c.Flags, // append built-in flags
			&Flag[bool]{Names: []string{"help", "h"}, Usage: "Show help", Value: &c.showHelp},
			&Flag[bool]{Names: []string{"version", "v"}, Usage: "Print the version", Value: &c.showVersion},
		)
	})
}

// Help generates and returns a formatted help message for the command.
func (c *Command) Help() string {
	c.init()

	const offset = "   " // indentation offset for formatting

	var b strings.Builder

	b.Grow(len(c.Description) + len(c.Name) + len(c.Version) + len(c.Flags)*64)

	// append the description if available
	if c.Description != "" {
		b.WriteString("Description:\n")
		b.WriteString(offset)
		b.WriteString(c.Description)
	}

	// append usage information
	if c.Name != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}

		b.WriteString("Usage:\n")
		b.WriteString(offset)
		b.WriteString(c.Name)

		if c.Usage != "" {
			b.WriteRune(' ')
			b.WriteString(c.Usage)
		}
	}

	// append version information
	if c.Version != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}

		b.WriteString("Version:\n")
		b.WriteString(offset)
		b.WriteString(c.Version)
	}

	// append flags if any exist
	if len(c.Flags) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}

		b.WriteString("Options:\n")

		var (
			longest               int // stores the length of the longest flag name for alignment
			flagNames, flagUsages = make([]string, len(c.Flags)), make([]string, len(c.Flags))
		)

		// iterate through flags to determine the longest name
		for i, f := range c.Flags {
			flagNames[i], flagUsages[i] = f.Help()

			if l := utf8.RuneCountInString(flagNames[i]); l > longest {
				longest = l
			}
		}

		// append flag information to the help message
		for i, flagName := range flagNames {
			if i > 0 {
				b.WriteRune('\n')
			}

			b.WriteString(offset)
			b.WriteString(flagName)

			// align flag descriptions
			for j := utf8.RuneCountInString(flagName); j < longest; j++ {
				b.WriteRune(' ')
			}

			b.WriteString("  ")
			b.WriteString(flagUsages[i])
		}
	}

	return b.String()
}

// Run executes the command with the provided arguments.
func (c *Command) Run(ctx context.Context, args []string) error { //nolint:contextcheck
	if ctx == nil { // nil ctx fallback: no parent context to inherit from
		ctx = context.Background()
	} else if err := ctx.Err(); err != nil {
		return err // do nothing if the context is already canceled
	}

	c.init()

	// create a new flag set for parsing command-line flags
	var set = flag.NewFlagSet(c.Name, flag.ContinueOnError)

	// suppress output from the standard flag library to avoid unnecessary messages
	set.SetOutput(io.Discard)

	// set default output if not defined
	if c.Output == nil {
		c.Output = os.Stdout
	}

	// register flags in the flag set
	for _, f := range c.Flags {
		f.Apply(set)
	}

	// parse command-line arguments
	if err := set.Parse(args); err != nil {
		// display help message in case of a parsing error
		if _, outErr := fmt.Fprintf(c.Output, "%s\n", c.Help()); outErr != nil {
			err = fmt.Errorf("%w: %w", outErr, err)
		}

		return err
	}

	// if help flag is set, print help message and exit (before flags validation and other actions)
	if c.showHelp {
		_, err := fmt.Fprintf(c.Output, "%s\n", c.Help())

		return err
	}

	// if version flag is set, print version information and exit
	if c.showVersion {
		var (
			runtimeVersion = runtime.Version()
			out            string
		)

		if c.Version != "" {
			out = fmt.Sprintf("%s (%s)\n", c.Version, runtimeVersion)
		} else {
			out = fmt.Sprintf("unknown (%s)\n", runtimeVersion)
		}

		_, err := fmt.Fprint(c.Output, out)

		return err
	}

	// validate and execute any flag-specific actions
	for _, f := range c.Flags {
		if !f.IsSet() {
			continue
		}

		// validate the flag before running its action
		if err := f.Validate(c); err != nil {
			return err
		}

		// run the flag's action if applicable
		if err := f.RunAction(c); err != nil {
			return err
		}
	}

	// execute the main command action if set
	if c.Action != nil {
		return c.Action(ctx, c, set.Args())
	}

	return nil
}
