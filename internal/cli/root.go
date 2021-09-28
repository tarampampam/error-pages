// Package cli contains CLI command handlers.
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tarampampam/error-pages/internal/checkers"
	buildCmd "github.com/tarampampam/error-pages/internal/cli/build"
	healthcheckCmd "github.com/tarampampam/error-pages/internal/cli/healthcheck"
	serveCmd "github.com/tarampampam/error-pages/internal/cli/serve"
	versionCmd "github.com/tarampampam/error-pages/internal/cli/version"
	"github.com/tarampampam/error-pages/internal/env"
	"github.com/tarampampam/error-pages/internal/logger"
	"github.com/tarampampam/error-pages/internal/version"
)

const configFileFlagName = "config-file"

// NewCommand creates root command.
func NewCommand(appName string) *cobra.Command { //nolint:funlen
	var (
		configFile string
		verbose    bool
		debug      bool
		logJSON    bool
	)

	ctx := context.Background() // main CLI context

	// create "default" logger (will be overwritten later with customized)
	log, err := logger.New(false, false, false)
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use: appName,
		PersistentPreRunE: func(c *cobra.Command, _ []string) error {
			_ = log.Sync() // sync previous logger instance

			customizedLog, e := logger.New(verbose, debug, logJSON)
			if e != nil {
				return e
			}

			*log = *customizedLog // override "default" logger with customized

			c.Flags().VisitAll(func(flag *pflag.Flag) {
				// flag was NOT defined using CLI (flags should have maximal priority)
				if !flag.Changed && flag.Name == configFileFlagName {
					if envConfigFile, exists := env.ConfigFilePath.Lookup(); exists && envConfigFile != "" {
						configFile = envConfigFile
					}
				}
			})

			return nil
		},
		PersistentPostRun: func(*cobra.Command, []string) {
			// error ignoring reasons:
			// - <https://github.com/uber-go/zap/issues/772>
			// - <https://github.com/uber-go/zap/issues/328>
			_ = log.Sync()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "debug output")
	cmd.PersistentFlags().BoolVarP(&logJSON, "log-json", "", false, "logs in JSON format")
	cmd.PersistentFlags().StringVarP(
		&configFile,
		configFileFlagName, "c",
		"./error-pages.yml",
		fmt.Sprintf("path to the config file [$%s]", env.ConfigFilePath),
	)

	cmd.AddCommand(
		versionCmd.NewCommand(version.Version()),
		healthcheckCmd.NewCommand(checkers.NewHealthChecker(ctx)),
		buildCmd.NewCommand(log, &configFile),
		serveCmd.NewCommand(ctx, log, &configFile),
	)

	return cmd
}
