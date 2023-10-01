package cli_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/cli"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	app := cli.NewApp("app")

	assert.NotEmpty(t, app.Flags)

	assert.NoError(t, app.Run(context.Background(), []string{"", "--log-level", "debug", "--log-format", "json"}))
}
