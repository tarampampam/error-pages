package cli

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	_ "github.com/urfave/cli-docs/v3" // required for `go generate` to work
	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/appmeta"
	"gh.tarampamp.am/error-pages/internal/cli/build"
	"gh.tarampamp.am/error-pages/internal/cli/healthcheck"
	"gh.tarampamp.am/error-pages/internal/cli/perftest"
	"gh.tarampamp.am/error-pages/internal/cli/serve"
	"gh.tarampamp.am/error-pages/internal/logger"
)

//go:generate go run update_readme.go

// NewApp creates a new console application.
func NewApp(appName string) *cli.Command { //nolint:funlen
	var (
		logLevelFlag = cli.StringFlag{
			Name:     "log-level",
			Value:    logger.InfoLevel.String(),
			Usage:    "logging level (" + strings.Join(logger.LevelStrings(), "/") + ")",
			Sources:  cli.EnvVars("LOG_LEVEL"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Validator: func(s string) error {
				if _, err := logger.ParseLevel(s); err != nil {
					return err
				}

				return nil
			},
		}

		logFormatFlag = cli.StringFlag{
			Name:     "log-format",
			Value:    logger.ConsoleFormat.String(),
			Usage:    "logging format (" + strings.Join(logger.FormatStrings(), "/") + ")",
			Sources:  cli.EnvVars("LOG_FORMAT"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Validator: func(s string) error {
				if _, err := logger.ParseFormat(s); err != nil {
					return err
				}

				return nil
			},
		}
	)

	// create a "default" logger (will be swapped later with customized)
	var log, _ = logger.New(logger.InfoLevel, logger.ConsoleFormat) // error will never occur

	return &cli.Command{
		Usage:   appName,
		Suggest: true,
		Before: func(ctx context.Context, c *cli.Command) error {
			var (
				logLevel, _  = logger.ParseLevel(c.String(logLevelFlag.Name))   // error ignored because the flag validates itself
				logFormat, _ = logger.ParseFormat(c.String(logFormatFlag.Name)) // --//--
			)

			configured, err := logger.New(logLevel, logFormat) // create a new logger instance
			if err != nil {
				return err
			}

			*log = *configured // swap the "default" logger with customized

			return nil
		},
		Commands: []*cli.Command{
			serve.NewCommand(log),
			build.NewCommand(log),
			healthcheck.NewCommand(log, healthcheck.NewHTTPHealthChecker()),
			perftest.NewCommand(log),
		},
		Version: fmt.Sprintf("%s (%s)", appmeta.Version(), runtime.Version()),
		Flags: []cli.Flag{ // global flags
			&logLevelFlag,
			&logFormatFlag,
		},
	}
}
