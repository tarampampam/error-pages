package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tarampampam/error-pages/internal/cli"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	app := cli.NewApp("app")

	assert.NotEmpty(t, app.Flags)

	assert.NoError(t, app.Run([]string{"", "--log-level", "debug", "--log-format", "json"}))
}
