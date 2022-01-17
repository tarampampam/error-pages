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
		cfg           *config.Config
	)

	cmd := &cobra.Command{
		Use:     "build <output-directory>",
		Aliases: []string{"b"},
		Short:   "Build the error pages",
		Args:    cobra.ExactArgs(1),
		PreRunE: func(*cobra.Command, []string) error {
			if configFile == nil {
				return errors.New("path to the config file is required for this command")
			}

			if c, err := config.FromYamlFile(*configFile); err != nil {
				return err
			} else {
				if err = c.Validate(); err != nil {
					return err
				}

				cfg = c
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("wrong arguments count")
			}

			return run(log, cfg, args[0], generateIndex)
		},
	}

	cmd.Flags().BoolVarP(
		&generateIndex,
		"index", "i",
		false,
		"generate index page",
	)

	return cmd
}

const (
	outHTMLFileExt   = ".html"
	outIndexFileName = "index"
	outFilePerm      = os.FileMode(0664)
	outDirPerm       = os.FileMode(0775)
)

func run(log *zap.Logger, cfg *config.Config, outDirectoryPath string, generateIndex bool) error {
	log.Info("loading templates")

	templates, tplLoadingErr := cfg.LoadTemplates()
	if tplLoadingErr != nil {
		return tplLoadingErr
	} else if len(templates) == 0 {
		return errors.New("no loaded templates")
	}

	log.Info("output directory preparing", zap.String("Path", outDirectoryPath))

	if err := createDirectory(outDirectoryPath); err != nil {
		return errors.Wrap(err, "cannot prepare output directory")
	}

	history := newBuildingHistory()

	for templateName, templateContent := range templates {
		log.Debug("template processing", zap.String("name", templateName))

		for pageCode, pageProperties := range cfg.Pages {
			if err := createDirectory(path.Join(outDirectoryPath, templateName)); err != nil {
				return err
			}

			var fileName = pageCode + outHTMLFileExt

			if err := os.WriteFile(
				path.Join(outDirectoryPath, templateName, fileName),
				tpl.Render(templateContent, tpl.Properties{
					Code:        pageCode,
					Message:     pageProperties.Message,
					Description: pageProperties.Description,
				}),
				outFilePerm,
			); err != nil {
				return err
			}

			history.Append(
				templateName,
				pageCode,
				cfg.Pages[pageCode].Message,
				path.Join(templateName, fileName),
			)
		}
	}

	if generateIndex {
		var filepath = path.Join(outDirectoryPath, outIndexFileName+outHTMLFileExt)

		log.Info("index file generation", zap.String("path", filepath))

		if err := history.WriteIndexFile(filepath, outFilePerm); err != nil {
			return err
		}
	}

	return nil
}

func createDirectory(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, outDirPerm)
		}

		return err
	}

	if !stat.IsDir() {
		return errors.New("is not a directory")
	}

	return nil
}
