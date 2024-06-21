package shared

import (
	"fmt"
	"net"

	"github.com/urfave/cli/v3"
)

var ListenAddrFlag = cli.StringFlag{
	Name:     "listen",
	Aliases:  []string{"l"},
	Usage:    "IP (v4 or v6) address to listen on",
	Value:    "0.0.0.0", // bind to all interfaces by default
	Sources:  cli.EnvVars("LISTEN_ADDR"),
	OnlyOnce: true,
	Required: true,
	Config:   cli.StringConfig{TrimSpace: true},
	Validator: func(ip string) error {
		if ip == "" {
			return fmt.Errorf("missing IP address")
		}

		if net.ParseIP(ip) == nil {
			return fmt.Errorf("wrong IP address [%s] for listening", ip)
		}

		return nil
	},
}

var ListenPortFlag = cli.UintFlag{
	Name:     "port",
	Aliases:  []string{"p"},
	Usage:    "TCP port number",
	Value:    8080, // default port number
	Sources:  cli.EnvVars("LISTEN_PORT"),
	OnlyOnce: true,
	Required: true,
	Validator: func(port uint64) error {
		if port == 0 || port > 65535 {
			return fmt.Errorf("wrong TCP port number [%d]", port)
		}

		return nil
	},
}
