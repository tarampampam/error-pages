package cli_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/cli"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestCommand_Help(t *testing.T) {
	t.Parallel()

	var builtInFlagsHelp = `Options:
   --help, -h     Show help
   --version, -v  Print the version`

	for name, tc := range map[string]struct {
		giveCommand *cli.Command
		wantHelp    string
	}{
		"empty": {
			giveCommand: &cli.Command{},
			wantHelp:    builtInFlagsHelp,
		},
		"with description": {
			giveCommand: &cli.Command{
				Description: "Some description here",
			},
			wantHelp: "Description:\n   Some description here\n\n" + builtInFlagsHelp,
		},
		"with name": {
			giveCommand: &cli.Command{
				Name: "some-name",
			},
			wantHelp: "Usage:\n   some-name\n\n" + builtInFlagsHelp,
		},
		"with name and usage": {
			giveCommand: &cli.Command{
				Name:  "some-name",
				Usage: "some-usage",
			},
			wantHelp: "Usage:\n   some-name some-usage\n\n" + builtInFlagsHelp,
		},
		"full": {
			giveCommand: &cli.Command{
				Name:        "some-name",
				Description: "Some description here",
				Usage:       "some-usage",
				Version:     "some-version",
				Flags: []cli.Flagger{
					&cli.Flag[string]{
						Names:   []string{"config-file", "c"},
						Usage:   "Path to the configuration file",
						EnvVars: []string{"CONFIG_FILE"},
					},
				},
			},
			wantHelp: `Description:
   Some description here

Usage:
   some-name some-usage

Version:
   some-version

Options:
   --config-file="…", -c="…"  Path to the configuration file [$CONFIG_FILE]
   --help, -h                 Show help
   --version, -v              Print the version`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantHelp, tc.giveCommand.Help())
		})
	}
}

func TestCommand_Run(t *testing.T) {
	t.Parallel()

	var ctx = context.Background()

	t.Run("canceled context", func(t *testing.T) {
		t.Parallel()

		var c = &cli.Command{}

		newCtx, cancel := context.WithCancel(ctx)
		cancel()

		assert.ErrorContains(t, c.Run(newCtx, nil), context.Canceled.Error())
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		var c = &cli.Command{}

		assert.NoError(t, c.Run(ctx, nil))
		assert.NoError(t, c.Run(nil, nil)) //nolint:contextcheck,staticcheck
	})

	t.Run("help (built-in flag)", func(t *testing.T) {
		t.Parallel()

		var (
			out      strings.Builder
			executed bool

			c = &cli.Command{
				Name:   "some-name",
				Output: &out,
				Action: func(context.Context, *cli.Command, []string) (_ error) { executed = true; return },
			}
		)

		for _, arg := range [...]string{"--help", "-h"} {
			assert.NoError(t, c.Run(ctx, []string{arg}))
			assert.Equal(t, false, executed) // should not execute the action
			assert.Equal(t, c.Help()+"\n", out.String())

			out.Reset()
		}
	})

	t.Run("version (built-in flag)", func(t *testing.T) {
		t.Parallel()

		var (
			out            strings.Builder
			runtimeVersion = runtime.Version()
			executed       bool

			c = &cli.Command{
				Name:    "some-name",
				Version: "some-version",
				Output:  &out,
				Action:  func(context.Context, *cli.Command, []string) (_ error) { executed = true; return },
			}
		)

		for _, arg := range [...]string{"--version", "-v"} {
			assert.NoError(t, c.Run(ctx, []string{arg}))
			assert.Equal(t, false, executed) // should not execute the action
			assert.Equal(t, fmt.Sprintf("%s (%s)\n", c.Version, runtimeVersion), out.String())

			out.Reset()
		}

		c.Version = "" // unset version

		for _, arg := range [...]string{"--version", "-v"} {
			assert.NoError(t, c.Run(ctx, []string{arg}))
			assert.Equal(t, false, executed) // should not execute the action
			assert.Equal(t, fmt.Sprintf("unknown (%s)\n", runtimeVersion), out.String())

			out.Reset()
		}
	})

	t.Run("custom flag action", func(t *testing.T) {
		t.Parallel()

		var (
			cmdActExecuted  bool
			flagActExecuted bool
			testErr         = errors.New("test error")

			c = &cli.Command{
				Name: "some-name",
				Flags: []cli.Flagger{
					&cli.Flag[bool]{
						Names:  []string{"custom-flag", "f"},
						Action: func(_ *cli.Command, _ bool) error { flagActExecuted = true; return testErr },
					},
				},
				Action: func(context.Context, *cli.Command, []string) error { cmdActExecuted = true; return nil },
			}
		)

		for _, arg := range [...]string{"--custom-flag", "-f"} {
			assert.ErrorContains(t, c.Run(ctx, []string{arg}), testErr.Error())
			assert.Equal(t, true, flagActExecuted)
			assert.Equal(t, false, cmdActExecuted)

			cmdActExecuted, flagActExecuted = false, false // reset
		}
	})

	t.Run("custom flag validation", func(t *testing.T) {
		t.Parallel()

		var (
			value   string
			testErr = errors.New("invalid value")

			c = &cli.Command{
				Name: "some-name",
				Flags: []cli.Flagger{
					&cli.Flag[string]{
						Names: []string{"custom-flag", "f"},
						Validator: func(_ *cli.Command, s string) error {
							if s == "valid" {
								return nil
							}

							return testErr
						},
						Value: &value,
					},
				},
			}
		)

		// valid value
		for _, args := range [...][]string{
			{"--custom-flag=valid"},
			{"--custom-flag", "valid"},
			{"-f=valid"},
			{"-f", "valid"},
		} {
			assert.NoError(t, c.Run(ctx, args))
			assert.Equal(t, "valid", value)

			value = "" // reset
		}

		// invalid value
		for _, args := range [...][]string{
			{"--custom-flag=invalid"},
			{"--custom-flag", "invalid"},
			{"-f=invalid"},
			{"-f", "invalid"},
		} {
			assert.Equal(t, testErr, c.Run(ctx, args))
			assert.Equal(t, "invalid", value) // the value is set anyway

			value = "" // reset
		}
	})

	t.Run("custom flag parsing error", func(t *testing.T) {
		t.Parallel()

		var (
			value int
			out   strings.Builder

			c = &cli.Command{
				Name:   "some-name",
				Output: &out,
				Flags: []cli.Flagger{
					&cli.Flag[bool]{
						Names: []string{"bar"},
					},
					&cli.Flag[int]{
						Names: []string{"custom-flag", "f"},
						Value: &value,
					},
				},
			}
		)

		for _, args := range [...][]string{
			{"--bar", "--custom-flag=foo"},
			{"--custom-flag", "foo"},
			{"-f=foo"},
			{"-f", "foo", "--bar"},
		} {
			assert.ErrorContains(t, c.Run(ctx, args), "invalid value")
			assert.Equal(t, c.Help()+"\n", out.String())

			out.Reset() // reset
		}
	})

	t.Run("command action", func(t *testing.T) {
		t.Parallel()

		var (
			executed bool

			c = &cli.Command{
				Name:   "some-name",
				Action: func(context.Context, *cli.Command, []string) error { executed = true; return nil },
			}
		)

		assert.NoError(t, c.Run(ctx, nil))
		assert.Equal(t, true, executed)
	})
}
