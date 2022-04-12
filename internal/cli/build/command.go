package build

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tarampampam/error-pages/internal/config"
	"github.com/tarampampam/error-pages/internal/tpl"
	"go.uber.org/zap"
)

// NewCommand creates `build` command.
func NewCommand(log *zap.Logger, configFile *string) *cobra.Command {
	var (
		generateIndex bool
		disableL10n   bool
		cfg           *config.Config
	)

	cmd := &cobra.Command{
		Use:     "build <output-directory>",
		Aliases: []string{"b"},
		Short:   "Build the error pages",
		Args:    cobra.ExactArgs(1),
		PreRunE: func(*cobra.Command, []string) (err error) {
			if configFile == nil {
				return errors.New("path to the config file is required for this command")
			}

			if cfg, err = config.FromYamlFile(*configFile); err != nil {
				return err
			}

			return
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("wrong arguments count")
			}

			return run(log, cfg, args[0], generateIndex, disableL10n)
		},
	}

	cmd.Flags().BoolVarP(
		&generateIndex,
		"index", "i",
		false,
		"generate index page",
	)

	cmd.Flags().BoolVarP(
		&disableL10n,
		"disable-l10n", "",
		false,
		"disable error pages localization",
	)

	return cmd
}

const (
	outHTMLFileExt   = ".html"
	outIndexFileName = "index"
	outFilePerm      = os.FileMode(0664)
	outDirPerm       = os.FileMode(0775)
)

func run(log *zap.Logger, cfg *config.Config, outDirectoryPath string, generateIndex, disableL10n bool) error { //nolint:funlen,lll
	if len(cfg.Templates) == 0 {
		return errors.New("no loaded templates")
	}

	log.Info("output directory preparing", zap.String("path", outDirectoryPath))

	if err := createDirectory(outDirectoryPath, outDirPerm); err != nil {
		return errors.Wrap(err, "cannot prepare output directory")
	}

	history, renderer := newBuildingHistory(), tpl.NewTemplateRenderer()
	defer func() { _ = renderer.Close() }()

	for _, template := range cfg.Templates {
		log.Debug("template processing", zap.String("name", template.Name()))

		for _, page := range cfg.Pages {
			if err := createDirectory(path.Join(outDirectoryPath, template.Name()), outDirPerm); err != nil {
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

func createDirectory(path string, perm os.FileMode) error {
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
