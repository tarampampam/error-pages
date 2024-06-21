package cli_test

import (
	"context"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/cli"
)

func TestNewApp(t *testing.T) {
	t.Parallel()

	app := cli.NewApp("appName")

	assert.NotEmpty(t, app.Flags)

	output := capturer.CaptureStdout(func() {
		assert.NoError(t, app.Run(context.Background(), []string{""}))
	})

	assert.NotEmpty(t, output)
}
