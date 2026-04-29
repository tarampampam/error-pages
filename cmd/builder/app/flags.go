package app

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gh.tarampamp.am/error-pages/v4/internal/cli"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/template/tploader"
)

func newCreateIndexFlag() cli.Flag[bool] {
	return cli.Flag[bool]{
		Names:   []string{"index"},
		Usage:   "Create an index.html file with links to all generated error pages",
		EnvVars: []string{"CREATE_INDEX"},
	}
}

func newTargetDirPath(def string) cli.Flag[string] {
	return cli.Flag[string]{
		Names:   []string{"out", "target-dir", "o"},
		Usage:   "Directory to place the built error pages",
		Default: def,
		EnvVars: []string{"OUT_DIR"},
		Validator: func(_ *cli.Command, dir string) error {
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
}

func newTemplateFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names:   []string{"template"},
		Usage:   "Custom template for error pages",
		EnvVars: []string{"TEMPLATE"},
		Validator: func(_ *cli.Command, src string) error {
			if tploader.IsURL(src) || tploader.IsFilePath(src) {
				// if it's a URL or file path, we will attempt to load it later, so just skip validation for now
				return nil
			}

			t, err := tpl.New(src)
			if err != nil {
				return fmt.Errorf("custom template parsing: %w", err)
			}

			if err = t.RenderTo(tpl.Data{}, io.Discard); err != nil {
				return fmt.Errorf("custom template rendering test: %w", err)
			}

			return nil
		},
	}
}
