package healthcheck_test

import (
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/tarampampam/error-pages/internal/cli/healthcheck"
)

type fakeChecker struct{ err error }

func (c *fakeChecker) Check(port uint16) error { return c.err }

func TestProperties(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})

	assert.Equal(t, "healthcheck", cmd.Name)
	assert.ElementsMatch(t, []string{"chk", "health", "check"}, cmd.Aliases)
	assert.NotNil(t, cmd.Action)
}

func TestCommandRun(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})

	assert.NoError(t, cmd.Run(cli.NewContext(cli.NewApp(), &flag.FlagSet{}, nil)))
}

func TestCommandRunFailed(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: errors.New("foo err")})

	assert.ErrorContains(t, cmd.Run(cli.NewContext(cli.NewApp(), &flag.FlagSet{}, nil)), "foo err")
}

func TestPortFlagWrongArgument(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})

	err := cmd.Run(
		cli.NewContext(cli.NewApp(), &flag.FlagSet{}, nil),
		"", "-p", "65536",
	)

	assert.ErrorContains(t, err, "port value out of range")
}
