package build

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/logger"
	appTemplate "gh.tarampamp.am/error-pages/internal/template"
)

//go:embed index.html
var indexHtml string

type command struct {
	c *cli.Command

	opt struct {
		createIndex      bool
		targetDirAbsPath string
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
			Name:     "index",
			Aliases:  []string{"i"},
			Usage:    "Generate index.html file with links to all error pages",
			Category: shared.CategoryBuild,
		}
		targetDirFlag = cli.StringFlag{
			Name:     "target-dir",
			Aliases:  []string{"out", "dir", "o"},
			Usage:    "Directory to put the built error pages into",
			Value:    ".", // current directory by default
			Config:   cli.StringConfig{TrimSpace: true},
			Category: shared.CategoryBuild,
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
			cmd.opt.targetDirAbsPath, _ = filepath.Abs(c.String(targetDirFlag.Name)) // an error checked by [os.Stat] validator

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

			log.Info("Building error pages",
				logger.String("targetDir", cmd.opt.targetDirAbsPath),
				logger.Strings("templates", cfg.Templates.Names()...),
				logger.Bool("index", cmd.opt.createIndex),
				logger.Bool("l10n", !cfg.L10n.Disable),
			)

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

func (cmd *command) Run( //nolint:funlen
	ctx context.Context,
	log *logger.Logger,
	cfg *config.Config,
) error {
	type historyItem struct{ Code, Message, RelativePath string }

	var history = make(map[string][]historyItem, len(cfg.Codes)*len(cfg.Templates)) // map[template_name]codes

	for templateName, templateContent := range cfg.Templates {
		log.Debug("Processing template", logger.String("name", templateName))

		for code, codeDescription := range cfg.Codes {
			if err := createDirectory(filepath.Join(cmd.opt.targetDirAbsPath, templateName)); err != nil {
				return fmt.Errorf("cannot create directory for template '%s': %w", templateName, err)
			}

			var codeAsUint, codeParsingErr = strconv.ParseUint(code, 10, 32)
			if codeParsingErr != nil {
				log.Warn("Cannot parse code", logger.String("code", code))

				continue
			}

			var outFilePath = path.Join(cmd.opt.targetDirAbsPath, templateName, code+".html")

			if content, renderErr := appTemplate.Render(templateContent, appTemplate.Props{
				Code:               uint16(codeAsUint),
				Message:            codeDescription.Message,
				Description:        codeDescription.Description,
				L10nDisabled:       cfg.L10n.Disable,
				ShowRequestDetails: false,
			}); renderErr == nil {
				if err := os.WriteFile(outFilePath, []byte(content), os.FileMode(0664)); err != nil { //nolint:mnd
					return err
				}
			} else {
				return fmt.Errorf("cannot render template '%s': %w", templateName, renderErr)
			}

			log.Debug("Page built", logger.String("template", templateName), logger.String("code", code))

			history[templateName] = append(history[templateName], historyItem{
				Code:         code,
				Message:      codeDescription.Message,
				RelativePath: "." + strings.TrimPrefix(outFilePath, cmd.opt.targetDirAbsPath), // to make it relative
			})
		}
	}

	if cmd.opt.createIndex {
		log.Debug("Creating the index file")

		for name := range history {
			slices.SortFunc(history[name], func(a, b historyItem) int { return strings.Compare(a.Code, b.Code) })
		}

		indexTpl, tplErr := template.New("index").Parse(indexHtml)
		if tplErr != nil {
			return tplErr
		}

		var buf strings.Builder

		if err := indexTpl.Execute(&buf, history); err != nil {
			return err
		}

		return os.WriteFile(
			filepath.Join(cmd.opt.targetDirAbsPath, "index.html"),
			[]byte(buf.String()),
			os.FileMode(0664), //nolint:mnd
		)
	}

	return nil
}

func createDirectory(path string) error {
	var stat, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, os.FileMode(0775)) //nolint:mnd
		}

		return err
	}

	if !stat.IsDir() {
		return errors.New("is not a directory")
	}

	return nil
}
