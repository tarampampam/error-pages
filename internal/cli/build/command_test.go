package build_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/zap"

	"gh.tarampamp.am/error-pages/internal/cli/build"
)

func TestNewCommand(t *testing.T) {
	defer goleak.VerifyNone(t)

	cmd := build.NewCommand(zap.NewNop())

	assert.NotEmpty(t, cmd.Flags)

	assert.Error(t, cmd.Run(
		context.Background(),
		[]string{},
	), "should fail with wrong arguments count")
}
