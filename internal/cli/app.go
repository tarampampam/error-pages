package cli

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"gh.tarampamp.am/error-pages/internal/checkers"
	"gh.tarampamp.am/error-pages/internal/cli/build"
	"gh.tarampamp.am/error-pages/internal/cli/healthcheck"
	"gh.tarampamp.am/error-pages/internal/cli/serve"
	"gh.tarampamp.am/error-pages/internal/env"
	"gh.tarampamp.am/error-pages/internal/logger"
	"gh.tarampamp.am/error-pages/internal/version"
)

// NewApp creates new console application.
func NewApp(appName string) *cli.App { //nolint:funlen
	const (
		logLevelFlagName  = "log-level"
		logFormatFlagName = "log-format"
		verboseFlagName   = "verbose"
		debugFlagName     = "debug"
		logJSONFlagName   = "log-json"

		defaultLogLevel  = logger.InfoLevel
		defaultLogFormat = logger.ConsoleFormat
	)

	// create "default" logger (will be overwritten later with customized)
	var log, _ = logger.New(defaultLogLevel, defaultLogFormat) // error will never occurs

	return &cli.App{
		Usage: appName,
		Before: func(c *cli.Context) (err error) {
			_ = log.Sync() // sync previous logger instance

			var logLevel, logFormat = defaultLogLevel, defaultLogFormat //nolint:ineffassign

			if c.Bool(verboseFlagName) || c.Bool(debugFlagName) {
				logLevel = logger.DebugLevel
			} else {
				// parse logging level
				if logLevel, err = logger.ParseLevel(c.String(logLevelFlagName)); err != nil {
					return err
				}
			}

			if c.Bool(logJSONFlagName) {
				logFormat = logger.JSONFormat
			} else {
				// parse logging format
				if logFormat, err = logger.ParseFormat(c.String(logFormatFlagName)); err != nil {
					return err
				}
			}

			configured, err := logger.New(logLevel, logFormat) // create new logger instance
			if err != nil {
				return err
			}

			*log = *configured // replace "default" logger with customized

			return nil
		},
		Commands: []*cli.Command{
			healthcheck.NewCommand(checkers.NewHealthChecker(context.TODO())),
			build.NewCommand(log),
			serve.NewCommand(log),
		},
		Version: fmt.Sprintf("%s (%s)", version.Version(), runtime.Version()),
		Flags: []cli.Flag{ // global flags
			&cli.BoolFlag{ // kept for backward compatibility
				Name:  verboseFlagName,
				Usage: "verbose output (DEPRECATED FLAG)",
			},
			&cli.BoolFlag{ // kept for backward compatibility
				Name:  debugFlagName,
				Usage: "debug output (DEPRECATED FLAG)",
			},
			&cli.BoolFlag{ // kept for backward compatibility
				Name:  logJSONFlagName,
				Usage: "logs in JSON format (DEPRECATED FLAG)",
			},
			&cli.StringFlag{
				Name:    logLevelFlagName,
				Value:   defaultLogLevel.String(),
				Usage:   "logging level (`" + strings.Join(logger.LevelStrings(), "/") + "`)",
				EnvVars: []string{env.LogLevel.String()},
			},
			&cli.StringFlag{
				Name:    logFormatFlagName,
				Value:   defaultLogFormat.String(),
				Usage:   "logging format (`" + strings.Join(logger.FormatStrings(), "/") + "`)",
				EnvVars: []string{env.LogFormat.String()},
			},
		},
	}
}
