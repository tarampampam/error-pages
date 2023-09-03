package healthcheck_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/cli/healthcheck"
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

	assert.NoError(t, cmd.Run(context.Background(), []string{}))
}

func TestCommandRunFailed(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: errors.New("foo err")})

	assert.ErrorContains(t, cmd.Run(context.Background(), []string{}), "foo err")
}

func TestPortFlagWrongArgument(t *testing.T) {
	cmd := healthcheck.NewCommand(&fakeChecker{err: nil})

	err := cmd.Run(context.Background(), []string{"", "-p", "65536"})

	assert.ErrorContains(t, err, "port value out of range")
}
