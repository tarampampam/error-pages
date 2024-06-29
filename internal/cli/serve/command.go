package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/config"
	appHttp "gh.tarampamp.am/error-pages/internal/http"
	"gh.tarampamp.am/error-pages/internal/logger"
)

type command struct {
	c *cli.Command

	opt struct {
		http struct { // our HTTP server
			addr string
			port uint16
			// readBufferSize uint
		}
	}
}

// NewCommand creates `serve` command.
func NewCommand(log *logger.Logger) *cli.Command { //nolint:funlen,gocognit,gocyclo
	var (
		cmd       command
		cfg       = config.New()
		env, trim = cli.EnvVars, cli.StringConfig{TrimSpace: true}
	)

	var (
		addrFlag        = shared.ListenAddrFlag
		portFlag        = shared.ListenPortFlag
		addTplFlag      = shared.AddTemplatesFlag
		disableTplFlag  = shared.DisableTemplateNamesFlag
		addCodeFlag     = shared.AddHTTPCodesFlag
		disableL10nFlag = shared.DisableL10nFlag
		jsonFormatFlag  = cli.StringFlag{
			Name:     "json-format",
			Usage:    "override the default error page response in JSON format (Go templates are supported; the error page will use this template if the client requests JSON content type)",
			Sources:  env("RESPONSE_JSON_FORMAT"),
			OnlyOnce: true,
			Config:   trim,
		}
		xmlFormatFlag = cli.StringFlag{
			Name:     "xml-format",
			Usage:    "override the default error page response in XML format (Go templates are supported; the error page will use this template if the client requests XML content type)",
			Sources:  env("RESPONSE_XML_FORMAT"),
			OnlyOnce: true,
			Config:   trim,
		}
		plainTextFormatFlag = cli.StringFlag{
			Name:     "plaintext-format",
			Usage:    "override the default error page response in plain text format (Go templates are supported; the error page will use this template if the client requests PlainText content type or does not specify any)",
			Sources:  env("RESPONSE_PLAINTEXT_FORMAT"),
			OnlyOnce: true,
			Config:   trim,
		}
		templateNameFlag = cli.StringFlag{
			Name:     "template-name",
			Aliases:  []string{"t"},
			Value:    cfg.TemplateName,
			Usage:    "name of the template to use for rendering error pages (builtin templates: " + strings.Join(cfg.Templates.Names(), ", ") + ")",
			Sources:  env("TEMPLATE_NAME"),
			OnlyOnce: true,
			Config:   trim,
		}
		defaultCodeToRenderFlag = cli.UintFlag{
			Name:    "default-error-page",
			Usage:   "the code of the default (index page, when a code is not specified) error page to render",
			Value:   uint64(cfg.DefaultCodeToRender),
			Sources: env("DEFAULT_ERROR_PAGE"),
			Validator: func(code uint64) error {
				if code > 999 { //nolint:mnd
					return fmt.Errorf("wrong HTTP code [%d] for the default error page", code)
				}

				return nil
			},
			OnlyOnce: true,
		}
		sendSameHTTPCodeFlag = cli.BoolFlag{
			Name: "send-same-http-code",
			Usage: "the HTTP response should have the same status code as the requested error page (by default, " +
				"every response with an error page will have a status code of 200)",
			Value:    cfg.RespondWithSameHTTPCode,
			Sources:  env("SEND_SAME_HTTP_CODE"),
			OnlyOnce: true,
		}
		showDetailsFlag = cli.BoolFlag{
			Name:     "show-details",
			Usage:    "show request details in the error page response (if supported by the template)",
			Value:    cfg.ShowDetails,
			Sources:  env("SHOW_DETAILS"),
			OnlyOnce: true,
		}
		proxyHeadersListFlag = cli.StringFlag{
			Name: "proxy-headers",
			Usage: "listed here HTTP headers will be proxied from the original request to the error page response " +
				"(comma-separated list)",
			Value:   strings.Join(cfg.ProxyHeaders, ","),
			Sources: env("PROXY_HTTP_HEADERS"),
			Validator: func(s string) error {
				for _, raw := range strings.Split(s, ",") {
					if clean := strings.TrimSpace(raw); strings.ContainsRune(clean, ' ') {
						return fmt.Errorf("whitespaces in the HTTP headers are not allowed: %s", clean)
					}
				}

				return nil
			},
			OnlyOnce: true,
			Config:   trim,
		}
		rotationModeFlag = cli.StringFlag{
			Name:     "rotation-mode",
			Value:    config.RotationModeDisabled.String(),
			Usage:    "templates automatic rotation mode (" + strings.Join(config.RotationModeStrings(), "/") + ")",
			Sources:  env("TEMPLATES_ROTATION_MODE"),
			OnlyOnce: true,
			Config:   trim,
			Validator: func(s string) error {
				if _, err := config.ParseRotationMode(s); err != nil {
					return err
				}

				return nil
			},
		}
	)

	addrFlag.Usage = "the HTTP server will listen on this IP (v4 or v6) address (set 127.0.0.1 for localhost, 0.0.0.0 to listen on all interfaces, or specify a custom IP)"
	portFlag.Usage = "the TPC port number for the HTTP server to listen on (0-65535)"

	disableL10nFlag.Value = cfg.L10n.Disable // set the default value depending on the configuration

	cmd.c = &cli.Command{
		Name:    "serve",
		Aliases: []string{"s", "server", "http"},
		Usage:   "Start HTTP server",
		Suggest: true,
		Action: func(ctx context.Context, c *cli.Command) error {
			cmd.opt.http.addr = c.String(addrFlag.Name)
			cmd.opt.http.port = uint16(c.Uint(portFlag.Name))
			cfg.L10n.Disable = c.Bool(disableL10nFlag.Name)
			cfg.DefaultCodeToRender = uint16(c.Uint(defaultCodeToRenderFlag.Name))
			cfg.RespondWithSameHTTPCode = c.Bool(sendSameHTTPCodeFlag.Name)
			cfg.RotationMode, _ = config.ParseRotationMode(c.String(rotationModeFlag.Name))
			cfg.ShowDetails = c.Bool(showDetailsFlag.Name)

			{ // override default JSON, XML, and PlainText formats
				if c.IsSet(jsonFormatFlag.Name) {
					cfg.Formats.JSON = strings.TrimSpace(c.String(jsonFormatFlag.Name))
				}

				if c.IsSet(xmlFormatFlag.Name) {
					cfg.Formats.XML = strings.TrimSpace(c.String(xmlFormatFlag.Name))
				}

				if c.IsSet(plainTextFormatFlag.Name) {
					cfg.Formats.PlainText = strings.TrimSpace(c.String(plainTextFormatFlag.Name))
				}
			}

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

			// set the list of HTTP headers we need to proxy from the incoming request to the error page response
			if c.IsSet(proxyHeadersListFlag.Name) {
				var m = make(map[string]struct{}) // map is used to avoid duplicates

				for _, header := range strings.Split(c.String(proxyHeadersListFlag.Name), ",") {
					m[http.CanonicalHeaderKey(strings.TrimSpace(header))] = struct{}{}
				}

				clear(cfg.ProxyHeaders) // clear the list before adding new headers

				for header := range m {
					cfg.ProxyHeaders = append(cfg.ProxyHeaders, header)
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

			// disable templates specified by the user
			if disable := c.StringSlice(disableTplFlag.Name); len(disable) > 0 {
				for _, templateName := range disable {
					if ok := cfg.Templates.Remove(templateName); ok {
						log.Info("Template disabled", logger.String("name", templateName))
					}
				}
			}

			// check if there are any templates available to render error pages
			if len(cfg.Templates.Names()) == 0 {
				return errors.New("no templates available to render error pages")
			}

			// if the rotation mode is set to random-on-startup, pick a random template (ignore the user-provided
			// template name)
			if cfg.RotationMode == config.RotationModeRandomOnStartup {
				cfg.TemplateName = cfg.Templates.RandomName()
			} else { // otherwise, use the user-provided template name
				cfg.TemplateName = c.String(templateNameFlag.Name)

				if !cfg.Templates.Has(cfg.TemplateName) {
					return fmt.Errorf(
						"template '%s' not found and cannot be used (available templates: %s)",
						cfg.TemplateName,
						cfg.Templates.Names(),
					)
				}
			}

			log.Debug("Configuration",
				logger.Strings("loaded templates", cfg.Templates.Names()...),
				logger.Strings("described HTTP codes", cfg.Codes.Codes()...),
				logger.String("JSON format", cfg.Formats.JSON),
				logger.String("XML format", cfg.Formats.XML),
				logger.String("plain text format", cfg.Formats.PlainText),
				logger.String("template name", cfg.TemplateName),
				logger.Bool("disable localization", cfg.L10n.Disable),
				logger.Uint16("default code to render", cfg.DefaultCodeToRender),
				logger.Bool("respond with the same HTTP code", cfg.RespondWithSameHTTPCode),
				logger.String("rotation mode", cfg.RotationMode.String()),
				logger.Bool("show details", cfg.ShowDetails),
				logger.Strings("proxy HTTP headers", cfg.ProxyHeaders...),
			)

			return cmd.Run(ctx, log, &cfg)
		},
		Flags: []cli.Flag{
			&addrFlag,
			&portFlag,
			&addTplFlag,
			&disableTplFlag,
			&addCodeFlag,
			&jsonFormatFlag,
			&xmlFormatFlag,
			&plainTextFormatFlag,
			&templateNameFlag,
			&disableL10nFlag,
			&defaultCodeToRenderFlag,
			&sendSameHTTPCodeFlag,
			&showDetailsFlag,
			&proxyHeadersListFlag,
			&rotationModeFlag,
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error { //nolint:funlen
	var srv = appHttp.NewServer(ctx, log)

	if err := srv.Register(cfg); err != nil {
		return err
	}

	var startingErrCh = make(chan error, 1) // channel for server starting error
	defer close(startingErrCh)

	// to track the frequency of each template's use, we send a simple GET request to the GoatCounter API
	// (https://goatcounter.com, https://github.com/arp242/goatcounter) to increment the counter. this service is
	// free and does not require an API key. no private data is sent, as shown in the URL below. this is necessary
	// to render a badge displaying the number of template usages on the error-pages repository README file :D
	//
	// badge code example:
	//	![Used times](https://img.shields.io/badge/dynamic/json?
	//		url=https%3A%2F%2Ferror-pages.goatcounter.com%2Fcounter%2F%2Fuse-template%2Flost-in-space.json
	//		&query=%24.count&label=Used%20times)
	//
	// if you wish, you may view the collected statistics at any time here - https://error-pages.goatcounter.com/
	go func() {
		var tpl = url.QueryEscape(cfg.TemplateName)

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(
			// https://www.goatcounter.com/help/pixel
			"https://error-pages.goatcounter.com/count?p=/use-template/%s&t=%s", tpl, tpl,
		), http.NoBody)
		if reqErr != nil {
			return
		}

		req.Header.Set("User-Agent", fmt.Sprintf("Mozilla/5.0 (error-pages, rnd:%d)", time.Now().UnixNano()))

		resp, respErr := (&http.Client{Timeout: 10 * time.Second}).Do(req) //nolint:mnd // don't care about the response
		if respErr != nil {
			log.Debug("Cannot send a request to increment the template usage counter", logger.Error(respErr))

			return
		} else if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		var now = time.Now()

		defer func() {
			log.Info("HTTP server stopped", logger.Duration("uptime", time.Since(now).Round(time.Millisecond)))
		}()

		log.Info("HTTP server starting",
			logger.String("addr", cmd.opt.http.addr),
			logger.Uint16("port", cmd.opt.http.port),
		)

		if err := srv.Start(cmd.opt.http.addr, cmd.opt.http.port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}(startingErrCh)

	// and wait for...
	select {
	case err := <-startingErrCh: // ..server starting error
		return err

	case <-ctx.Done(): // ..or context cancellation
		const shutdownTimeout = 5 * time.Second

		log.Info("HTTP server stopping", logger.Duration("with timeout", shutdownTimeout))

		if err := srv.Stop(shutdownTimeout); err != nil { //nolint:contextcheck
			return err
		}
	}

	return nil
}
