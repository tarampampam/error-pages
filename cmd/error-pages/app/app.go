package app

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
	"gh.tarampamp.am/error-pages/v4/internal/cli"
	"gh.tarampamp.am/error-pages/v4/internal/cli/shared"
	"gh.tarampamp.am/error-pages/v4/internal/codes"
	"gh.tarampamp.am/error-pages/v4/internal/errgroup"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver"
	"gh.tarampamp.am/error-pages/v4/internal/logger"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/template/tploader"
	"gh.tarampamp.am/error-pages/v4/templates"
)

//go:generate go run ./generate/readme.go -out ../../../docs/CLI.md

// App represents the CLI application with its command and options.
type App struct {
	cmd cli.Command

	opt struct {
		http struct {
			addr string
			port uint
		}
		errorPages struct {
			defaultCodeToRender uint
			sendSameHTTPCode    bool
			showDetails         bool
			proxyHeaders        []string
			disableBuiltInCodes bool
			addHTTPCodes        map[string]codes.Description
			templateName        string
			rotationMode        tpl.RotationMode
			customTemplates     struct {
				html, json, xml, text string
			}
			l10nDisabled bool
		}
	}
}

// NewApp initializes a new CLI application instance.
func NewApp(name string) *App { //nolint:funlen
	app := App{
		cmd: cli.Command{
			Name:        name,
			Description: "Start the HTTP server to serve the error pages",
			Version:     appmeta.Version(),
		},
	}

	allTemplateNames := slices.Collect(maps.Keys(templates.BuiltInHTML()))
	slices.Sort(allTemplateNames)

	app.opt.http.addr = "0.0.0.0" // bind to all interfaces by default
	app.opt.http.port = 8080
	app.opt.errorPages.defaultCodeToRender = uint(http.StatusNotFound)
	app.opt.errorPages.proxyHeaders = []string{"X-Request-Id", "X-Trace-Id", "X-Correlation-Id", "X-Amzn-Trace-Id"}
	app.opt.errorPages.templateName = templates.HTMLTemplateNameAppDown
	app.opt.errorPages.rotationMode = tpl.RotationModeDisabled

	var (
		logLevelFlag            = newLogLevelFlag()
		logFormatFlag           = newLogFormatFlag()
		httpAddrFlag            = newHTTPAddrFlag(app.opt.http.addr)
		httpPortFlag            = newHTTPPortFlag(app.opt.http.port)
		defaultCodeToRenderFlag = newDefaultCodeToRenderFlag(app.opt.errorPages.defaultCodeToRender)
		sendSameHTTPCodeFlag    = newSendSameHTTPCodeFlag()
		showDetailsFlag         = newShowDetailsFlag()
		proxyHeadersListFlag    = newProxyHeadersListFlag(app.opt.errorPages.proxyHeaders)
		disableBuiltInCodesFlag = shared.NewDisableBuiltInCodesFlag()
		addHTTPCodesFlag        = shared.NewAddHTTPCodesFlag()
		templateNameFlag        = newTemplateNameFlag(allTemplateNames, app.opt.errorPages.templateName)
		rotationModeFlag        = newRotationModeFlag(app.opt.errorPages.rotationMode)
		htmlTemplateFlag        = newHTMLTemplateFlag()
		jsonTemplateFlag        = newJSONTemplateFlag()
		xmlTemplateFlag         = newXMLTemplateFlag()
		textTemplateFlag        = newPlainTextTemplateFlag()
		disableL10nFlag         = shared.NewDisableL10nFlag()
	)

	app.cmd.Flags = []cli.Flagger{
		&logLevelFlag,
		&logFormatFlag,
		&httpAddrFlag,
		&httpPortFlag,
		&defaultCodeToRenderFlag,
		&sendSameHTTPCodeFlag,
		&showDetailsFlag,
		&proxyHeadersListFlag,
		&disableBuiltInCodesFlag,
		&addHTTPCodesFlag,
		&templateNameFlag,
		&rotationModeFlag,
		&htmlTemplateFlag,
		&jsonTemplateFlag,
		&xmlTemplateFlag,
		&textTemplateFlag,
		&disableL10nFlag,
	}

	app.cmd.Action = func(ctx context.Context, _ *cli.Command, _ []string) error {
		var (
			logLevel, _  = logger.ParseLevel(*logLevelFlag.Value)   //nolint:errcheck // because the flag validates itself
			logFormat, _ = logger.ParseFormat(*logFormatFlag.Value) //nolint:errcheck // format flag validates itself
		)

		log, logErr := logger.New(logLevel, logFormat)
		if logErr != nil {
			return logErr
		}

		ctx = logger.With(ctx, log)

		setIfFlagIsSet(&app.opt.http.addr, httpAddrFlag)
		setIfFlagIsSet(&app.opt.http.port, httpPortFlag)
		setIfFlagIsSet(&app.opt.errorPages.defaultCodeToRender, defaultCodeToRenderFlag)
		setIfFlagIsSet(&app.opt.errorPages.sendSameHTTPCode, sendSameHTTPCodeFlag)
		setIfFlagIsSet(&app.opt.errorPages.showDetails, showDetailsFlag)
		setIfFlagIsSet(&app.opt.errorPages.disableBuiltInCodes, disableBuiltInCodesFlag)

		if proxyHeadersListFlag.Value != nil && proxyHeadersListFlag.IsSet() {
			app.opt.errorPages.proxyHeaders = splitProxyHeadersList(*proxyHeadersListFlag.Value)
		}

		slices.Sort(app.opt.errorPages.proxyHeaders)

		if addHTTPCodesFlag.Value != nil && addHTTPCodesFlag.IsSet() {
			if parsed, err := shared.ParseAddHTTPCodes(*addHTTPCodesFlag.Value); err == nil {
				app.opt.errorPages.addHTTPCodes = parsed
			}
		}

		setIfFlagIsSet(&app.opt.errorPages.templateName, templateNameFlag)

		if rotationModeFlag.Value != nil && rotationModeFlag.IsSet() {
			app.opt.errorPages.rotationMode = tpl.RotationMode(*rotationModeFlag.Value)
		}

		setIfFlagIsSet(&app.opt.errorPages.customTemplates.html, htmlTemplateFlag)
		setIfFlagIsSet(&app.opt.errorPages.customTemplates.json, jsonTemplateFlag)
		setIfFlagIsSet(&app.opt.errorPages.customTemplates.xml, xmlTemplateFlag)
		setIfFlagIsSet(&app.opt.errorPages.customTemplates.text, textTemplateFlag)
		setIfFlagIsSet(&app.opt.errorPages.l10nDisabled, disableL10nFlag)

		// load custom templates concurrently if specified
		if err := app.loadTemplates(ctx); err != nil {
			log.Error("Failed to load custom templates", logger.Error(err))

			return errors.New("failed to load custom templates")
		}

		if err := app.run(ctx, log); err != nil {
			log.Error("HTTP server failed", logger.Error(err))

			return errors.New("http server failed")
		}

		return nil
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

// loadTemplates loads custom templates concurrently if they are specified in the options and appear to be from a
// valid source (URL or file path), and does nothing otherwise.
func (a *App) loadTemplates(ctx context.Context) error {
	ct := &a.opt.errorPages.customTemplates
	eg, _ := errgroup.New(ctx)

	if src := ct.html; src != "" {
		eg.Go(func(ctx context.Context) error {
			t, err := tploader.LoadTemplateContent(ctx, src)
			if err != nil {
				return fmt.Errorf("load HTML template: %w", err)
			}

			ct.html = t

			return nil
		})
	}

	if src := ct.json; src != "" {
		eg.Go(func(ctx context.Context) error {
			t, err := tploader.LoadTemplateContent(ctx, src)
			if err != nil {
				return fmt.Errorf("load JSON template: %w", err)
			}

			ct.json = t

			return nil
		})
	}

	if src := ct.xml; src != "" {
		eg.Go(func(ctx context.Context) error {
			t, err := tploader.LoadTemplateContent(ctx, src)
			if err != nil {
				return fmt.Errorf("load XML template: %w", err)
			}

			ct.xml = t

			return nil
		})
	}

	if src := ct.text; src != "" {
		eg.Go(func(ctx context.Context) error {
			t, err := tploader.LoadTemplateContent(ctx, src)
			if err != nil {
				return fmt.Errorf("load plain text template: %w", err)
			}

			ct.text = t

			return nil
		})
	}

	return eg.Wait()
}

// Run starts the CLI command execution.
func (a *App) Run(ctx context.Context, args []string) error { return a.cmd.Run(ctx, args) }

// run opens the TCP listener, starts the HTTP server, and blocks until the context is canceled or the server fails.
func (a *App) run(ctx context.Context, log *logger.Logger) error {
	log.Info("Opening TCP port",
		logger.String("addr", a.opt.http.addr),
		logger.Uint64("port", uint64(a.opt.http.port)),
	)

	ln, lnErr := (&net.ListenConfig{}).Listen(ctx, "tcp", net.JoinHostPort(
		a.opt.http.addr,
		strconv.Itoa(int(a.opt.http.port)), //nolint:gosec // port is validated to be in range 1-65535
	))
	if lnErr != nil {
		return fmt.Errorf("listen http: %w", lnErr)
	}

	defer func() { _ = ln.Close() }() // just in case, although http.Server should take care of it when shutting down

	httpCodes := codes.New(a.opt.errorPages.disableBuiltInCodes)

	// after this, we CAN'T modify httpCodes anymore, because it used concurrently
	maps.Copy(httpCodes, a.opt.errorPages.addHTTPCodes)

	templater, tErr := tpl.NewTemplates(
		tpl.WithCustomHTMLTemplate(a.opt.errorPages.customTemplates.html),
		tpl.WithCustomJSONTemplate(a.opt.errorPages.customTemplates.json),
		tpl.WithCustomXMLTemplate(a.opt.errorPages.customTemplates.xml),
		tpl.WithCustomPlainTextTemplate(a.opt.errorPages.customTemplates.text),
		tpl.WithHTMLTemplateName(a.opt.errorPages.templateName),
		tpl.WithRotationMode(a.opt.errorPages.rotationMode),
	)
	if tErr != nil {
		return fmt.Errorf("initialize templates: %w", tErr)
	}

	server := httpserver.New(
		httpserver.NewHandler(
			log,
			uint16(a.opt.errorPages.defaultCodeToRender), //nolint:gosec // validated to be in range 1-65535
			a.opt.errorPages.sendSameHTTPCode,
			a.opt.errorPages.proxyHeaders,
			httpCodes.Find,
			templater.Get,
			a.opt.errorPages.showDetails,
			a.opt.errorPages.l10nDisabled,
		),
		httpserver.WithErrorLog(logger.NewStdLog(log, logger.ErrorLevel)),
	)

	log.Info("Server configuration",
		logger.Strings("http_codes", httpCodes.Codes()...),
		logger.Bool("custom_html_template", strings.TrimSpace(a.opt.errorPages.customTemplates.html) != ""),
		logger.Bool("custom_json_template", strings.TrimSpace(a.opt.errorPages.customTemplates.json) != ""),
		logger.Bool("custom_xml_template", strings.TrimSpace(a.opt.errorPages.customTemplates.xml) != ""),
		logger.Bool("custom_text_template", strings.TrimSpace(a.opt.errorPages.customTemplates.text) != ""),
		logger.String("template_name", a.opt.errorPages.templateName),
		logger.String("rotation_mode", string(a.opt.errorPages.rotationMode)),
		logger.Uint64("default_error_page", uint64(a.opt.errorPages.defaultCodeToRender)),
		logger.Bool("send_same_http_code", a.opt.errorPages.sendSameHTTPCode),
		logger.Bool("show_details", a.opt.errorPages.showDetails),
		logger.Strings("proxy_headers", a.opt.errorPages.proxyHeaders...),
		logger.Bool("l10n_disabled", a.opt.errorPages.l10nDisabled),
	)

	now := time.Now()

	defer func() { log.Info("HTTP server stopped", logger.Duration("uptime", time.Since(now))) }()

	log.Info("HTTP server started", logger.String("addr", ln.Addr().String()))

	// since Serve() is blocking, we run it in the main goroutine and rely on context cancellation to stop it gracefully
	// when needed - this way we don't need to handle signals and shutdown logic here, and the server will take care
	// of it internally
	return server.Serve(ctx, ln)
}
