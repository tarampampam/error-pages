package serve

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/config"
)

type command struct {
	c *cli.Command

	opt struct{}
}

// NewCommand creates `serve` command.
func NewCommand(log *zap.Logger) *cli.Command { //nolint:funlen
	var cmd command

	var (
		portFlag    = shared.ListenPortFlag
		addrFlag    = shared.ListenAddrFlag
		addTplFlag  = shared.AddTemplateFlag
		addCodeFlag = shared.AddHTTPCodeFlag

		jsonFormatFlag = cli.StringFlag{
			Name:     "json-format",
			Usage:    "override the default error page response in JSON format (Go templates are supported)",
			Sources:  cli.EnvVars("RESPONSE_JSON_FORMAT"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
		}

		xmlFormatFlag = cli.StringFlag{
			Name:     "xml-format",
			Usage:    "override the default error page response in XML format (Go templates are supported)",
			Sources:  cli.EnvVars("RESPONSE_XML_FORMAT"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
		}
	)

	cmd.c = &cli.Command{
		Name:    "serve",
		Aliases: []string{"s", "server", "http"},
		Usage:   "Start HTTP server",
		Suggest: true,
		Action: func(ctx context.Context, c *cli.Command) error {
			var cfg = config.New()

			if add := c.StringSlice(addTplFlag.Name); len(add) > 0 { // add templates from files to the config
				for _, templatePath := range add {
					if addedName, err := cfg.Templates.AddFromFile(templatePath); err != nil {
						return fmt.Errorf("cannot add template from file %s: %w", templatePath, err)
					} else {
						log.Info("Template added",
							zap.String("name", addedName),
							zap.String("path", templatePath),
						)
					}
				}
			}

			if add := c.StringMap(addCodeFlag.Name); len(add) > 0 { // add custom HTTP codes
				for code, msgAndDesc := range add {
					var (
						parts = strings.SplitN(msgAndDesc, "/", 2) //nolint:mnd
						desc  config.CodeDescription
					)

					if len(parts) > 0 {
						desc.Message = strings.TrimSpace(parts[0])
					}

					if len(parts) > 1 {
						desc.Description = strings.TrimSpace(parts[1])
					}

					cfg.Codes[code] = desc

					log.Info("HTTP code added",
						zap.String("code", code),
						zap.String("message", desc.Message),
						zap.String("description", desc.Description),
					)
				}
			}

			{ // override default JSON and XML formats
				if c.IsSet(jsonFormatFlag.Name) {
					cfg.Formats.JSON = c.String(jsonFormatFlag.Name)
				}

				if c.IsSet(xmlFormatFlag.Name) {
					cfg.Formats.XML = c.String(xmlFormatFlag.Name)
				}
			}

			log.Debug("Configuration",
				zap.Strings("loaded templates", cfg.Templates.Names()),
				zap.Strings("described HTTP codes", cfg.Codes.Codes()),
				zap.String("JSON format", cfg.Formats.JSON),
				zap.String("XML format", cfg.Formats.XML),
			)

			return cmd.Run(ctx, log, &cfg)
		},
		Flags: []cli.Flag{
			&portFlag,
			&addrFlag,
			&addTplFlag,
			&addCodeFlag,
			&jsonFormatFlag,
			&xmlFormatFlag,
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(ctx context.Context, log *zap.Logger, cfg *config.Config) error {
	return nil // TODO: implement
}
