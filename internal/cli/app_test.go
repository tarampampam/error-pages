package cli_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/cli"
)

func TestNewApp(t *testing.T) {
	t.Parallel()

	app := cli.NewApp("appName")

	assert.NoError(t, app.Run(context.Background(), []string{""}))
}
