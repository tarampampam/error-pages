// Package healthcheck contains CLI `healthcheck` command implementation.
package healthcheck

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tarampampam/error-pages/internal/env"
)

type checker interface {
	Check(port uint16) error
}

const portFlagName = "port"

// NewCommand creates `healthcheck` command.
func NewCommand(checker checker) *cobra.Command {
	var port uint16

	cmd := &cobra.Command{
		Use:     "healthcheck",
		Aliases: []string{"chk", "health", "check"},
		Short:   "Health checker for the HTTP server. Use case - docker healthcheck",
		PreRunE: func(c *cobra.Command, _ []string) (lastErr error) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				// flag was NOT defined using CLI (flags should have maximal priority)
				if !flag.Changed && flag.Name == portFlagName {
					if envPort, exists := env.ListenPort.Lookup(); exists && envPort != "" {
						if p, err := strconv.ParseUint(envPort, 10, 16); err == nil { //nolint:gomnd
							port = uint16(p)
						} else {
							lastErr = fmt.Errorf("wrong TCP port environment variable [%s] value", envPort)
						}
					}
				}
			})

			return lastErr
		},
		RunE: func(*cobra.Command, []string) error {
			return checker.Check(port)
		},
	}

	cmd.Flags().Uint16VarP(
		&port,
		portFlagName,
		"p",
		8080, //nolint:gomnd // must be same as default serve `--port` flag value
		fmt.Sprintf("TCP port number [$%s]", env.ListenPort),
	)

	return cmd
}
