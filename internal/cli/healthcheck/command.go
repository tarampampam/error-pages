// Package healthcheck contains CLI `healthcheck` command implementation.
package healthcheck

import (
	"errors"
	"math"

	"github.com/urfave/cli/v2"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
)

type checker interface {
	Check(port uint16) error
}

// NewCommand creates `healthcheck` command.
func NewCommand(checker checker) *cli.Command {
	return &cli.Command{
		Name:    "healthcheck",
		Aliases: []string{"chk", "health", "check"},
		Usage:   "Health checker for the HTTP server. Use case - docker healthcheck",
		Action: func(c *cli.Context) error {
			var port = c.Uint(shared.ListenPortFlag.Name)

			if port <= 0 || port > math.MaxUint16 {
				return errors.New("port value out of range")
			}

			return checker.Check(uint16(port))
		},
		Flags: []cli.Flag{
			shared.ListenPortFlag,
		},
	}
}
