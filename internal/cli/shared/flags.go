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

var ListenAddrFlag = &cli.StringFlag{ //nolint:gochecknoglobals
	Name:    "listen",
	Aliases: []string{"l"},
	Usage:   "IP address to Listen on",
	Value:   "0.0.0.0",
	EnvVars: []string{env.ListenAddr.String()},
}

var ListenPortFlag = &cli.UintFlag{ //nolint:gochecknoglobals
	Name:    "port",
	Aliases: []string{"p"},
	Usage:   "TCP port number",
	Value:   8080, //nolint:gomnd
	EnvVars: []string{env.ListenPort.String()},
}
