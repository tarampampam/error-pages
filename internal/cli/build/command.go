package build

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/logger"
)

type command struct {
	c *cli.Command

	opt struct {
		createIndex bool
		targetDir   string
	}
}

// NewCommand creates `build` command.
func NewCommand(log *logger.Logger) *cli.Command { //nolint:funlen,gocognit
	var (
		cmd command
		cfg = config.New()

		addTplFlag      = shared.AddTemplatesFlag
		disableTplFlag  = shared.DisableTemplateNamesFlag
		addCodeFlag     = shared.AddHTTPCodesFlag
		disableL10nFlag = shared.DisableL10nFlag
		createIndexFlag = cli.BoolFlag{
			Name:    "index",
			Aliases: []string{"i"},
			Usage:   "generate index.html file with links to all error pages",
		}
		targetDirFlag = cli.StringFlag{
			Name:     "target-dir",
			Aliases:  []string{"out", "dir", "o"},
			Usage:    "directory to put the built error pages into",
			Value:    ".", // current directory by default
			Config:   cli.StringConfig{TrimSpace: true},
			OnlyOnce: true,
			Validator: func(dir string) error {
				if dir == "" {
					return errors.New("missing target directory")
				}

				if stat, err := os.Stat(dir); err != nil {
					return fmt.Errorf("cannot access the target directory '%s': %w", dir, err)
				} else if !stat.IsDir() {
					return fmt.Errorf("'%s' is not a directory", dir)
				}

				return nil
			},
		}
	)

	disableL10nFlag.Value = cfg.L10n.Disable // set the default value depending on the configuration

	cmd.c = &cli.Command{
		Name:    "build",
		Aliases: []string{"b"},
		Usage:   "Build the static error pages and put them into a specified directory",
		Action: func(ctx context.Context, c *cli.Command) error {
			cfg.L10n.Disable = c.Bool(disableL10nFlag.Name)
			cmd.opt.createIndex = c.Bool(createIndexFlag.Name)
			cmd.opt.targetDir, _ = filepath.Abs(c.String(targetDirFlag.Name)) // an error checked by [os.Stat] validator

			// add templates from files to the configuration
			if add := c.StringSlice(addTplFlag.Name); len(add) > 0 {
				for _, templatePath := range add {
					if addedName, err := cfg.Templates.AddFromFile(templatePath); err != nil {
						return fmt.Errorf("cannot add template from file %s: %w", templatePath, err)
					} else {
						log.Info("Template added",
							logger.String("name", addedName),
							logger.String("path", templatePath),
						)
					}
				}
			}

			// disable templates specified by the user
			if disable := c.StringSlice(disableTplFlag.Name); len(disable) > 0 {
				for _, templateName := range disable {
					if ok := cfg.Templates.Remove(templateName); ok {
						log.Info("Template disabled", logger.String("name", templateName))
					}
				}
			}

			// add custom HTTP codes to the configuration
			if add := c.StringMap(addCodeFlag.Name); len(add) > 0 {
				for code, desc := range shared.ParseHTTPCodes(add) {
					cfg.Codes[code] = desc

					log.Info("HTTP code added",
						logger.String("code", code),
						logger.String("message", desc.Message),
						logger.String("description", desc.Description),
					)
				}
			}

			if len(cfg.Templates) == 0 {
				return errors.New("no templates specified")
			}

			return cmd.Run(ctx, log, &cfg)
		},
		Flags: []cli.Flag{
			&addTplFlag,
			&disableTplFlag,
			&addCodeFlag,
			&disableL10nFlag,
			&createIndexFlag,
			&targetDirFlag,
		},
	}

	return cmd.c
}

func (cmd *command) Run(
	ctx context.Context,
	log *logger.Logger,
	cfg *config.Config,
) error {
	return nil
}
