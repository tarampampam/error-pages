package build_test

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"go.uber.org/goleak"
	"go.uber.org/zap"

	"github.com/tarampampam/error-pages/internal/cli/build"
)

func TestNewCommand(t *testing.T) {
	defer goleak.VerifyNone(t)

	cmd := build.NewCommand(zap.NewNop())

	assert.NotEmpty(t, cmd.Flags)

	assert.Error(t, cmd.Run(
		cli.NewContext(cli.NewApp(), &flag.FlagSet{}, nil),
		"",
	), "should fail because of missing external services")
}
