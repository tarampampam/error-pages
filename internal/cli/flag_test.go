package cli_test

import (
	"errors"
	"flag"
	"io"
	"os"
	"testing"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/cli"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/random"
)

func TestFlag_IsSet(t *testing.T) {
	t.Parallel()

	assert.Equal(t, false, (&cli.Flag[string]{
		// no value
	}).IsSet())

	assert.Equal(t, false, (&cli.Flag[int]{
		Value:        new(int),
		ValueSetFrom: cli.FlagValueSourceNone,
	}).IsSet())

	assert.Equal(t, false, (&cli.Flag[int]{
		Value:        new(int),
		ValueSetFrom: cli.FlagValueSourceDefault,
	}).IsSet())

	var intValue = 42

	assert.Equal(t, false, (&cli.Flag[int]{
		Value:        &intValue,
		ValueSetFrom: cli.FlagValueSourceFlag,
		Default:      intValue,
	}).IsSet())

	assert.Equal(t, true, (&cli.Flag[bool]{
		Value:        new(bool),
		ValueSetFrom: cli.FlagValueSourceFlag,
		Default:      true,
	}).IsSet())
}

func TestFlag_Help(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveFlag             cli.Flag[string]
		wantNames, wantUsage string
	}{
		"empty": {
			giveFlag:  cli.Flag[string]{},
			wantNames: "",
			wantUsage: "",
		},
		"single long name": {
			giveFlag:  cli.Flag[string]{Names: []string{"name"}},
			wantNames: `--name="…"`,
			wantUsage: "",
		},
		"single short name": {
			giveFlag:  cli.Flag[string]{Names: []string{"n"}},
			wantNames: `-n="…"`,
			wantUsage: "",
		},
		"multiple names": {
			giveFlag:  cli.Flag[string]{Names: []string{"name", "n"}},
			wantNames: `--name="…", -n="…"`,
			wantUsage: "",
		},
		"with usage": {
			giveFlag:  cli.Flag[string]{Usage: "usage\nfoo"},
			wantNames: "",
			wantUsage: "usage\nfoo",
		},
		"with default": {
			giveFlag:  cli.Flag[string]{Default: "default"},
			wantNames: "",
			wantUsage: "(default: default)",
		},
		"with env vars": {
			giveFlag:  cli.Flag[string]{EnvVars: []string{"ENV1", "ENV2"}},
			wantNames: "",
			wantUsage: "[$ENV1, $ENV2]",
		},
		"full": {
			giveFlag: cli.Flag[string]{
				Names:   []string{"name", "n"},
				Usage:   "usage\nfoo",
				Default: "default",
				EnvVars: []string{"ENV1", "ENV2"},
			},
			wantNames: `--name="…", -n="…"`,
			wantUsage: "usage\nfoo (default: default) [$ENV1, $ENV2]",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotNames, gotUsage := tc.giveFlag.Help()

			assert.Equal(t, tc.wantNames, gotNames)
			assert.Equal(t, tc.wantUsage, gotUsage)
		})
	}

	t.Run("bool", func(t *testing.T) {
		t.Run("default true", func(t *testing.T) {
			gotNames, gotUsage := (&cli.Flag[bool]{
				Names:   []string{"name", "n"},
				Usage:   "usage\nfoo",
				Default: true,
				EnvVars: []string{"ENV1", "ENV2"},
			}).Help()

			assert.Equal(t, `--name, -n`, gotNames)
			assert.Equal(t, "usage\nfoo (default: true) [$ENV1, $ENV2]", gotUsage)
		})

		t.Run("default false", func(t *testing.T) {
			gotNames, gotUsage := (&cli.Flag[bool]{
				Names:   []string{"name"},
				Usage:   "usage\nfoo",
				Default: false,
				EnvVars: []string{"ENV1", "ENV2"},
			}).Help()

			assert.Equal(t, `--name`, gotNames)
			assert.Equal(t, "usage\nfoo [$ENV1, $ENV2]", gotUsage)
		})
	})
}

func TestFlag_Apply(t *testing.T) {
	t.Parallel()

	t.Run("bool, default", func(t *testing.T) {
		t.Parallel()

		var (
			val bool
			f   = &cli.Flag[bool]{Names: []string{"test"}, Value: &val, Default: true}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, true, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("bool, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val bool
			f   = &cli.Flag[bool]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test"}))
		assert.Equal(t, true, val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("bool, env", func(t *testing.T) {
		t.Parallel()

		var (
			envName = setRandomEnv(t, "True")
			val     bool
			f       = &cli.Flag[bool]{Names: []string{"test"}, Value: &val, EnvVars: []string{envName}}
			set     = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, true, val)
		assert.Equal(t, cli.FlagValueSourceEnv, f.ValueSetFrom)
	})

	t.Run("bool, wrong env", func(t *testing.T) {
		t.Parallel()

		var (
			envName = setRandomEnv(t, "<invalid+Boolean-Value]")
			val     bool
			f       = &cli.Flag[bool]{Names: []string{"test"}, EnvVars: []string{envName}}
			set     = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, false, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("int, default", func(t *testing.T) {
		t.Parallel()

		var (
			val int
			f   = &cli.Flag[int]{Names: []string{"test"}, Value: &val, Default: 42}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, 42, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("int, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val int
			f   = &cli.Flag[int]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test=42"}))
		assert.Equal(t, 42, val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("int, wrong flag", func(t *testing.T) {
		t.Parallel()

		var (
			val int
			f   = &cli.Flag[int]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.ContinueOnError)
		)

		f.Apply(set)

		assert.ErrorContains(t, set.Parse([]string{"--test=foo"}), "must contain only digits with an optional leading")
		assert.Equal(t, 0, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("int, env", func(t *testing.T) {
		t.Parallel()

		var (
			envName = setRandomEnv(t, "42")
			val     int
			f       = &cli.Flag[int]{Names: []string{"test"}, Value: &val, EnvVars: []string{envName}}
			set     = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, 42, val)
		assert.Equal(t, cli.FlagValueSourceEnv, f.ValueSetFrom)
	})

	t.Run("int, wrong env", func(t *testing.T) {
		t.Parallel()

		var (
			envName = setRandomEnv(t, "forty-two")
			val     int
			f       = &cli.Flag[int]{Names: []string{"test"}, EnvVars: []string{envName}}
			set     = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, 0, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("int64, default", func(t *testing.T) {
		t.Parallel()

		var (
			val int64
			f   = &cli.Flag[int64]{Names: []string{"test"}, Value: &val, Default: 42}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, int64(42), val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("int64, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val int64
			f   = &cli.Flag[int64]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test=-42"}))
		assert.Equal(t, int64(-42), val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("int64, wrong flag", func(t *testing.T) {
		t.Parallel()

		var (
			val int64
			f   = &cli.Flag[int64]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.ContinueOnError)
		)

		f.Apply(set)

		assert.ErrorContains(t, set.Parse([]string{"--test=foo"}), "must contain only digits with an optional leading")
		assert.Equal(t, int64(0), val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("string, default", func(t *testing.T) {
		t.Parallel()

		var (
			val string
			f   = &cli.Flag[string]{Names: []string{"test"}, Value: &val, Default: "default"}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, "default", val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("string, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val string
			f   = &cli.Flag[string]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test=foo"}))
		assert.Equal(t, "foo", val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("string, env", func(t *testing.T) {
		t.Parallel()

		var (
			envVal  = random.String(10)
			envName = setRandomEnv(t, envVal)

			val string
			f   = &cli.Flag[string]{Names: []string{"test"}, Value: &val, EnvVars: []string{envName}}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, envVal, val)
		assert.Equal(t, cli.FlagValueSourceEnv, f.ValueSetFrom)
	})

	t.Run("uint, default", func(t *testing.T) {
		t.Parallel()

		var (
			val uint
			f   = &cli.Flag[uint]{Names: []string{"test"}, Value: &val, Default: 42}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, uint(42), val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("uint, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val uint
			f   = &cli.Flag[uint]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test=42"}))
		assert.Equal(t, uint(42), val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("uint, wrong flag", func(t *testing.T) {
		t.Parallel()

		var (
			val uint
			f   = &cli.Flag[uint]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.ContinueOnError)
		)

		f.Apply(set)

		assert.ErrorContains(t, set.Parse([]string{"--test=foo"}), "must contain only digits")
		assert.Equal(t, uint(0), val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("uint64, default", func(t *testing.T) {
		t.Parallel()

		var (
			val uint64
			f   = &cli.Flag[uint64]{Names: []string{"test"}, Value: &val, Default: 42}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, uint64(42), val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("uint64, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val uint64
			f   = &cli.Flag[uint64]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test=42"}))
		assert.Equal(t, uint64(42), val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("uint64, wrong flag", func(t *testing.T) {
		t.Parallel()

		var (
			val uint64
			f   = &cli.Flag[uint64]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.ContinueOnError)
		)

		f.Apply(set)

		assert.ErrorContains(t, set.Parse([]string{"--test=foo"}), "must contain only digits")
		assert.Equal(t, uint64(0), val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("float64, default", func(t *testing.T) {
		t.Parallel()

		var (
			val float64
			f   = &cli.Flag[float64]{Names: []string{"test"}, Value: &val, Default: 42.42}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, 42.42, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("float64, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val float64
			f   = &cli.Flag[float64]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test=42.42"}))
		assert.Equal(t, 42.42, val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("float64, wrong flag", func(t *testing.T) {
		t.Parallel()

		var (
			val float64
			f   = &cli.Flag[float64]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.ContinueOnError)
		)

		f.Apply(set)

		assert.ErrorContains(t, set.Parse([]string{"--test=foo"}), "must contain only digits with an optional decimal")
		assert.Equal(t, 0.0, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("time.Duration, default", func(t *testing.T) {
		t.Parallel()

		var (
			val time.Duration
			f   = &cli.Flag[time.Duration]{Names: []string{"test"}, Value: &val, Default: time.Second}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, time.Second, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("time.Duration, flag", func(t *testing.T) {
		t.Parallel()

		var (
			val time.Duration
			f   = &cli.Flag[time.Duration]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse([]string{"--test=42s"}))
		assert.Equal(t, 42*time.Second, val)
		assert.Equal(t, cli.FlagValueSourceFlag, f.ValueSetFrom)
	})

	t.Run("time.Duration, wrong flag", func(t *testing.T) {
		t.Parallel()

		var (
			val time.Duration
			f   = &cli.Flag[time.Duration]{Names: []string{"test"}, Value: &val}
			set = newFlagSet(flag.ContinueOnError)
		)

		f.Apply(set)

		assert.ErrorContains(t, set.Parse([]string{"--test=foo"}), "must be a valid Go duration string")
		assert.Equal(t, 0*time.Second, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})

	t.Run("time.Duration, env", func(t *testing.T) {
		t.Parallel()

		var (
			envVal  = "42s"
			envName = setRandomEnv(t, envVal)

			val time.Duration
			f   = &cli.Flag[time.Duration]{Names: []string{"test"}, Value: &val, EnvVars: []string{envName}}
			set = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, 42*time.Second, val)
		assert.Equal(t, cli.FlagValueSourceEnv, f.ValueSetFrom)
	})

	t.Run("time.Duration, wrong env", func(t *testing.T) {
		t.Parallel()

		var (
			envName = setRandomEnv(t, "forty-two")
			val     time.Duration
			f       = &cli.Flag[time.Duration]{Names: []string{"test"}, EnvVars: []string{envName}}
			set     = newFlagSet(flag.PanicOnError)
		)

		f.Apply(set)

		assert.NoError(t, set.Parse(nil))
		assert.Equal(t, 0*time.Second, val)
		assert.Equal(t, cli.FlagValueSourceDefault, f.ValueSetFrom)
	})
}

func TestFlag_Validate(t *testing.T) {
	t.Parallel()

	t.Run("nil validator", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, (&cli.Flag[string]{Names: []string{"test"}}).Validate(&cli.Command{}))
	})

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()

		var (
			executed bool

			f = &cli.Flag[string]{
				Names: []string{"test"},
				Validator: func(command *cli.Command, s string) error {
					executed = true

					return nil
				},
			}
		)

		assert.ErrorContains(t, f.Validate(&cli.Command{}), "flag value is nil")
		assert.Equal(t, false, executed)
	})

	t.Run("validator error", func(t *testing.T) {
		t.Parallel()

		var (
			executed bool
			val      = "str"

			f = &cli.Flag[string]{
				Names: []string{"test"},
				Value: &val,
				Validator: func(command *cli.Command, s string) error {
					assert.Equal(t, val, s)

					executed = true

					return errors.New("test error")
				},
			}
		)

		assert.ErrorContains(t, f.Validate(&cli.Command{}), "test error")
		assert.Equal(t, true, executed)
	})
}

func TestFlag_RunAction(t *testing.T) {
	t.Parallel()

	t.Run("nil action", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, (&cli.Flag[string]{Names: []string{"test"}}).RunAction(&cli.Command{}))
	})

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()

		var (
			executed bool

			f = &cli.Flag[string]{
				Names: []string{"test"},
				Action: func(command *cli.Command, s string) error {
					executed = true

					return nil
				},
			}
		)

		assert.ErrorContains(t, f.RunAction(&cli.Command{}), "flag value is nil")
		assert.Equal(t, false, executed)
	})

	t.Run("action error", func(t *testing.T) {
		t.Parallel()

		var (
			executed bool
			val      = "str"

			f = &cli.Flag[string]{
				Names: []string{"test"},
				Value: &val,
				Action: func(command *cli.Command, s string) error {
					assert.Equal(t, val, s)

					executed = true

					return errors.New("test error")
				},
			}
		)

		assert.ErrorContains(t, f.RunAction(&cli.Command{}), "test error")
		assert.Equal(t, true, executed)
	})
}

func newFlagSet(eh flag.ErrorHandling) *flag.FlagSet {
	var set = flag.NewFlagSet("test", eh)

	set.SetOutput(io.Discard)

	return set
}

func setRandomEnv(t *testing.T, value string) (envName string) {
	t.Helper()

	envName = random.String(10)

	assert.NoError(t, os.Setenv(envName, value)) //nolint:usetesting // using [t.Setenv] is not possible due t.Parallel()

	t.Cleanup(func() { assert.NoError(t, os.Unsetenv(envName)) })

	return envName
}
