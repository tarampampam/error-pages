// Package version contains CLI `version` command implementation.
package version

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// NewCommand creates `version` command.
func NewCommand(ver string) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v", "ver"},
		Short:   "Display application version",
		RunE: func(*cobra.Command, []string) (err error) {
			_, err = fmt.Fprintf(os.Stdout, "app version:\t%s (%s)\n", ver, runtime.Version())

			return
		},
	}
}
