package app

import (
	"context"
	_ "embed"
	"fmt"
	htmltpl "html/template"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
	"gh.tarampamp.am/error-pages/v4/internal/cli"
	"gh.tarampamp.am/error-pages/v4/internal/cli/shared"
	"gh.tarampamp.am/error-pages/v4/internal/codes"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/template/tploader"
	"gh.tarampamp.am/error-pages/v4/templates"
)

//go:generate go run ./generate/readme.go -out ../../../docs/CLI.md

//go:embed index.tpl.html
var indexTpl string

// App represents the CLI application with its command and options.
type App struct {
	cmd cli.Command

	opt struct {
		createIndex         bool
		targetDirAbsPath    string
		disableBuiltInCodes bool
		addHTTPCodes        map[string]codes.Description
		customTemplate      string
		l10nDisabled        bool
	}
}

// NewApp initializes a new CLI application instance.
func NewApp(name string) *App {
	app := App{
		cmd: cli.Command{
			Name: name,
			Description: "Build the static error pages and place them in the specified directory. If no custom " +
				"template is provided, the built-in one will be used.",
			Version: appmeta.Version(),
		},
	}

	var (
		createIndexFlag         = newCreateIndexFlag()
		targetDirPath           = newTargetDirPath(".")
		disableBuiltInCodesFlag = shared.NewDisableBuiltInCodesFlag()
		addHTTPCodesFlag        = shared.NewAddHTTPCodesFlag()
		templateFlag            = newTemplateFlag()
		disableL10nFlag         = shared.NewDisableL10nFlag()
	)

	app.cmd.Flags = []cli.Flagger{
		&createIndexFlag,
		&targetDirPath,
		&disableBuiltInCodesFlag,
		&addHTTPCodesFlag,
		&templateFlag,
		&disableL10nFlag,
	}

	app.cmd.Action = func(ctx context.Context, _ *cli.Command, _ []string) error {
		setIfFlagIsSet(&app.opt.createIndex, createIndexFlag)
		setIfFlagIsSet(&app.opt.targetDirAbsPath, targetDirPath)

		app.opt.targetDirAbsPath, _ = filepath.Abs(app.opt.targetDirAbsPath) //nolint:errcheck // checked by validator

		setIfFlagIsSet(&app.opt.disableBuiltInCodes, disableBuiltInCodesFlag)

		if addHTTPCodesFlag.Value != nil && addHTTPCodesFlag.IsSet() {
			if parsed, err := shared.ParseAddHTTPCodes(*addHTTPCodesFlag.Value); err == nil {
				app.opt.addHTTPCodes = parsed
			}
		}

		setIfFlagIsSet(&app.opt.customTemplate, templateFlag)

		// load custom template content if a source is provided (either URL, file path, or raw template string)
		if src := app.opt.customTemplate; src != "" {
			t, err := tploader.LoadTemplateContent(ctx, src)
			if err != nil {
				return fmt.Errorf("load custom template: %w", err)
			}

			app.opt.customTemplate = t
		}

		setIfFlagIsSet(&app.opt.l10nDisabled, disableL10nFlag)

		return app.run(ctx)
	}

	return &app
}

// Help returns the help message.
func (a *App) Help() string { return a.cmd.Help() }

// setIfFlagIsSet copies source's value into target only if the flag was explicitly provided by the user, not just
// defaulted. This matters because Flag.Value is always non-nil (set to default) after parsing, so IsSet is the only
// reliable way to know whether the user actually supplied the value.
func setIfFlagIsSet[T cli.FlagType](target *T, source cli.Flag[T]) {
	if target == nil || source.Value == nil || !source.IsSet() {
		return
	}

	*target = *source.Value
}

// Run starts the CLI command execution.
func (a *App) Run(ctx context.Context, args []string) error { return a.cmd.Run(ctx, args) }

type historyItem struct{ Code, Message, RelativePath string }

const fileMode os.FileMode = 0o664

func (a *App) run(_ context.Context) error {
	httpCodes := codes.New(a.opt.disableBuiltInCodes)
	maps.Copy(httpCodes, a.opt.addHTTPCodes)

	var history = make(map[string][]historyItem)

	if a.opt.customTemplate != "" {
		if err := a.renderCustomTemplate(httpCodes, history); err != nil {
			return err
		}
	} else {
		if err := a.renderBuiltInTemplates(httpCodes, history); err != nil {
			return err
		}
	}

	if !a.opt.createIndex || len(history) == 0 {
		return nil
	}

	for name := range history {
		slices.SortFunc(history[name], func(a, b historyItem) int { return strings.Compare(a.Code, b.Code) })
	}

	idxTpl, idxErr := htmltpl.New("index").Parse(indexTpl)
	if idxErr != nil {
		return fmt.Errorf("parse index template: %w", idxErr)
	}

	var buf strings.Builder

	if err := idxTpl.Execute(&buf, history); err != nil {
		return fmt.Errorf("render index template: %w", err)
	}

	indexPath := filepath.Join(a.opt.targetDirAbsPath, "index.html")

	return os.WriteFile(indexPath, []byte(buf.String()), fileMode)
}

// renderCustomTemplate renders all numeric HTTP codes using the custom template and writes them directly into
// the target directory as {code}.html files.
func (a *App) renderCustomTemplate(httpCodes codes.Codes, history map[string][]historyItem) error {
	t, err := tpl.New(a.opt.customTemplate)
	if err != nil {
		return fmt.Errorf("parse custom template: %w", err)
	}

	for _, codeStr := range httpCodes.Codes() {
		codeUint, parseErr := strconv.ParseUint(codeStr, 10, 16)
		if parseErr != nil {
			continue // skip wildcard codes like "4xx"
		}

		code := uint16(codeUint)

		desc, ok := httpCodes.Find(code)
		if !ok {
			continue
		}

		content, renderErr := t.Render(tpl.Data{
			StatusCode:  code,
			Message:     desc.Short,
			Description: desc.Full,
			Config:      tpl.Config{L10nDisabled: a.opt.l10nDisabled},
		})
		if renderErr != nil {
			return fmt.Errorf("render custom template for code %s: %w", codeStr, renderErr)
		}

		outPath := filepath.Join(a.opt.targetDirAbsPath, codeStr+".html")

		if wErr := os.WriteFile(outPath, content, fileMode); wErr != nil {
			return fmt.Errorf("write %s: %w", outPath, wErr)
		}

		history["custom"] = append(history["custom"], historyItem{
			Code:         codeStr,
			Message:      desc.Short,
			RelativePath: "." + strings.TrimPrefix(outPath, a.opt.targetDirAbsPath),
		})
	}

	return nil
}

// renderBuiltInTemplates renders all numeric HTTP codes for every built-in HTML template and writes them into
// per-template subdirectories as {templateName}/{code}.html files.
func (a *App) renderBuiltInTemplates(httpCodes codes.Codes, history map[string][]historyItem) error { //nolint:cyclop
	builtIn := templates.BuiltInHTML()

	allTemplateNames := slices.Collect(maps.Keys(builtIn))
	slices.Sort(allTemplateNames)

	for _, templateName := range allTemplateNames {
		subDir := filepath.Join(a.opt.targetDirAbsPath, templateName)

		if mkErr := os.MkdirAll(subDir, 0o775); mkErr != nil { //nolint:mnd
			return fmt.Errorf("create directory for template %q: %w", templateName, mkErr)
		}

		t, tplErr := tpl.New(builtIn[templateName])
		if tplErr != nil {
			return fmt.Errorf("parse built-in template %q: %w", templateName, tplErr)
		}

		for _, codeStr := range httpCodes.Codes() {
			codeUint, parseErr := strconv.ParseUint(codeStr, 10, 16)
			if parseErr != nil {
				continue // skip wildcard codes like "4xx"
			}

			code := uint16(codeUint)

			desc, ok := httpCodes.Find(code)
			if !ok {
				continue
			}

			content, renderErr := t.Render(tpl.Data{
				StatusCode:  code,
				Message:     desc.Short,
				Description: desc.Full,
				Config:      tpl.Config{L10nDisabled: a.opt.l10nDisabled},
			})
			if renderErr != nil {
				return fmt.Errorf("render template %q for code %s: %w", templateName, codeStr, renderErr)
			}

			outPath := filepath.Join(subDir, codeStr+".html")

			if wErr := os.WriteFile(outPath, content, fileMode); wErr != nil {
				return fmt.Errorf("write %s: %w", outPath, wErr)
			}

			history[templateName] = append(history[templateName], historyItem{
				Code:         codeStr,
				Message:      desc.Short,
				RelativePath: "." + strings.TrimPrefix(outPath, a.opt.targetDirAbsPath),
			})
		}
	}

	return nil
}
