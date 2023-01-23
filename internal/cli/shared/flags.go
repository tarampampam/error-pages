package shared

import (
	"github.com/urfave/cli/v2"

	"github.com/tarampampam/error-pages/internal/env"
)

var ConfigFileFlag = &cli.StringFlag{ //nolint:gochecknoglobals
	Name:    "config-file",
	Aliases: []string{"c"},
	Usage:   "path to the config file (yaml)",
	Value:   "./error-pages.yml",
	EnvVars: []string{env.ConfigFilePath.String()},
}
