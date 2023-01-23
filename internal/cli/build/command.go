package build

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/tarampampam/error-pages/internal/cli/shared"
	"github.com/tarampampam/error-pages/internal/config"
	"github.com/tarampampam/error-pages/internal/tpl"
)

type command struct {
	c *cli.Command
}

// NewCommand creates `build` command.
func NewCommand(log *zap.Logger) *cli.Command {
	var cmd = command{}

	const (
		generateIndexFlagName = "index"
		disableL10nFlagName   = "disable-l10n"
	)

	cmd.c = &cli.Command{
		Usage:       "build <output-directory>",
		Aliases:     []string{"b"},
		Description: "Build the error pages",
		Action: func(c *cli.Context) error {
			cfg, cfgErr := config.FromYamlFile(c.String(shared.ConfigFileFlag.Name))
			if cfgErr != nil {
				return cfgErr
			}

			if c.Args().Len() != 1 {
				return errors.New("wrong arguments count")
			}

			return cmd.Run(log, cfg, c.Args().First(), c.Bool(generateIndexFlagName), c.Bool(disableL10nFlagName))
		},
		Flags: []cli.Flag{ // global flags
			&cli.BoolFlag{
				Name:    generateIndexFlagName,
				Aliases: []string{"i"},
				Usage:   "generate index page",
			},
			&cli.BoolFlag{
				Name:  disableL10nFlagName,
				Usage: "disable error pages localization",
			},
			shared.ConfigFileFlag,
		},
	}

	return cmd.c
}

const (
	outHTMLFileExt   = ".html"
	outIndexFileName = "index"
	outFilePerm      = os.FileMode(0664)
	outDirPerm       = os.FileMode(0775)
)

func (cmd *command) Run(log *zap.Logger, cfg *config.Config, outDirectoryPath string, generateIndex, disableL10n bool) error { //nolint:funlen,lll
	if len(cfg.Templates) == 0 {
		return errors.New("no loaded templates")
	}

	log.Info("output directory preparing", zap.String("path", outDirectoryPath))

	if err := cmd.createDirectory(outDirectoryPath, outDirPerm); err != nil {
		return errors.Wrap(err, "cannot prepare output directory")
	}

	history, renderer := newBuildingHistory(), tpl.NewTemplateRenderer()
	defer func() { _ = renderer.Close() }()

	for _, template := range cfg.Templates {
		log.Debug("template processing", zap.String("name", template.Name()))

		for _, page := range cfg.Pages {
			if err := cmd.createDirectory(path.Join(outDirectoryPath, template.Name()), outDirPerm); err != nil {
				return err
			}

			var (
				fileName = page.Code() + outHTMLFileExt
				filePath = path.Join(outDirectoryPath, template.Name(), fileName)
			)

			content, renderingErr := renderer.Render(template.Content(), tpl.Properties{
				Code:               page.Code(),
				Message:            page.Message(),
				Description:        page.Description(),
				ShowRequestDetails: false,
				L10nDisabled:       disableL10n,
			})
			if renderingErr != nil {
				return renderingErr
			}

			if err := os.WriteFile(filePath, content, outFilePerm); err != nil {
				return err
			}

			log.Debug("page rendered", zap.String("path", filePath))

			if generateIndex {
				history.Append(
					template.Name(),
					page.Code(),
					page.Message(),
					path.Join(template.Name(), fileName),
				)
			}
		}
	}

	if generateIndex {
		var filepath = path.Join(outDirectoryPath, outIndexFileName+outHTMLFileExt)

		log.Info("index file generation", zap.String("path", filepath))

		if err := history.WriteIndexFile(filepath, outFilePerm); err != nil {
			return err
		}
	}

	log.Info("job is done")

	return nil
}

func (cmd *command) createDirectory(path string, perm os.FileMode) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, perm)
		}

		return err
	}

	if !stat.IsDir() {
		return errors.New("is not a directory")
	}

	return nil
}
