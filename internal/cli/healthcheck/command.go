package healthcheck

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/logger"
)

type checker interface {
	Check(ctx context.Context, baseURL string) error
}

// NewCommand creates `healthcheck` command.
func NewCommand(_ *logger.Logger, checker checker) *cli.Command {
	var portFlag = shared.ListenPortFlag

	return &cli.Command{
		Name:    "healthcheck",
		Aliases: []string{"chk", "health", "check"},
		Usage:   "Health checker for the HTTP server. The use case - docker health check",
		Action: func(ctx context.Context, c *cli.Command) error {
			return checker.Check(ctx, fmt.Sprintf("http://127.0.0.1:%d", c.Uint(portFlag.Name)))
		},
		Flags: []cli.Flag{
			&portFlag,
		},
	}
}
