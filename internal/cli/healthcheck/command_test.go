package healthcheck_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/cli/healthcheck"
	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	var cmd = healthcheck.NewCommand(logger.NewNop(), nil)

	assert.Equal(t, "healthcheck", cmd.Name)
	assert.Equal(t, []string{"chk", "health", "check"}, cmd.Aliases)
}

type fakeHealthChecker struct {
	t           *testing.T
	wantAddress string
	giveErr     error
}

func (m *fakeHealthChecker) Check(_ context.Context, addr string) error {
	assert.Equal(m.t, m.wantAddress, addr)

	return m.giveErr
}

func TestCommand_RunSuccess(t *testing.T) {
	t.Parallel()

	var cmd = healthcheck.NewCommand(logger.NewNop(), &fakeHealthChecker{
		t:           t,
		wantAddress: "http://127.0.0.1:1234",
	})

	require.NoError(t, cmd.Run(context.Background(), []string{"", "--port", "1234"}))
}

func TestCommand_RunFail(t *testing.T) {
	t.Parallel()

	cmd := healthcheck.NewCommand(logger.NewNop(), &fakeHealthChecker{
		t:           t,
		wantAddress: "http://127.0.0.1:4321",
		giveErr:     assert.AnError,
	})

	assert.ErrorIs(t,
		cmd.Run(context.Background(), []string{"", "--port", "4321"}),
		assert.AnError,
	)
}
