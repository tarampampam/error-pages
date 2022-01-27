package healthcheck_test

import (
	"errors"
	"os"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/cli/healthcheck"
)

type fakeChecker struct{ err error }

func (c *fakeChecker) Check(port uint16) error { return c.err }

func TestProperties(t *testing.T) {
	t.Parallel()

	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})

	assert.Equal(t, "healthcheck", cmd.Use)
	assert.ElementsMatch(t, []string{"chk", "health", "check"}, cmd.Aliases)
	assert.NotNil(t, cmd.RunE)
}

func TestCommandRun(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})
	cmd.SetArgs([]string{})

	output := capturer.CaptureOutput(func() {
		assert.NoError(t, cmd.Execute())
	})

	assert.Empty(t, output)
}

func TestCommandRunFailed(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: errors.New("foo err")})
	cmd.SetArgs([]string{})

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "foo err")
}

func TestPortFlagWrongArgument(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})
	cmd.SetArgs([]string{"-p", "65536"}) // 65535 is max

	var executed bool

	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true

		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "invalid argument")
	assert.Contains(t, output, "65536")
	assert.Contains(t, output, "value out of range")
	assert.False(t, executed)
}

func TestPortFlagWrongEnvValue(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})
	cmd.SetArgs([]string{})

	assert.NoError(t, os.Setenv("LISTEN_PORT", "65536")) // 65535 is max

	defer func() { assert.NoError(t, os.Unsetenv("LISTEN_PORT")) }()

	var executed bool

	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true

		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "wrong TCP port")
	assert.Contains(t, output, "environment variable")
	assert.Contains(t, output, "65536")
	assert.False(t, executed)
}
