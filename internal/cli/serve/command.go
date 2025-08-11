package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
			addr           string
			port           uint16
			readBufferSize uint
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
		addrFlag                = shared.ListenAddrFlag
		portFlag                = shared.ListenPortFlag
		addTplFlag              = shared.AddTemplatesFlag
		disableTplFlag          = shared.DisableTemplateNamesFlag
		addCodeFlag             = shared.AddHTTPCodesFlag
		disableL10nFlag         = shared.DisableL10nFlag
		disableMinificationFlag = shared.DisableMinificationFlag
		jsonFormatFlag          = cli.StringFlag{
			Name: "json-format",
			Usage: "Override the default error page response in JSON format (Go templates are supported; the error " +
				"page will use this template if the client requests JSON content type)",
			Sources:  env("RESPONSE_JSON_FORMAT"),
			Category: shared.CategoryFormats,
			OnlyOnce: true,
			Config:   trim,
		}
		xmlFormatFlag = cli.StringFlag{
			Name: "xml-format",
			Usage: "Override the default error page response in XML format (Go templates are supported; the error " +
				"page will use this template if the client requests XML content type)",
			Sources:  env("RESPONSE_XML_FORMAT"),
			Category: shared.CategoryFormats,
			OnlyOnce: true,
			Config:   trim,
		}
		plainTextFormatFlag = cli.StringFlag{
			Name: "plaintext-format",
			Usage: "Override the default error page response in plain text format (Go templates are supported; the " +
				"error page will use this template if the client requests plain text content type or does not specify any)",
			Sources:  env("RESPONSE_PLAINTEXT_FORMAT"),
			Category: shared.CategoryFormats,
			OnlyOnce: true,
			Config:   trim,
		}
		templateNameFlag = cli.StringFlag{
			Name:    "template-name",
			Aliases: []string{"t", "template", "theme"},
			Value:   cfg.TemplateName,
			Usage: "Name of the template to use for rendering error pages (built-in templates: " +
				strings.Join(cfg.Templates.Names(), ", ") + ")",
			Sources:  env("TEMPLATE_NAME"),
			Category: shared.CategoryTemplates,
			OnlyOnce: true,
			Config:   trim,
		}
		defaultCodeToRenderFlag = cli.UintFlag{
			Name:     "default-error-page",
			Usage:    "The code of the default (index page, when a code is not specified) error page to render",
			Value:    uint(cfg.DefaultCodeToRender),
			Sources:  env("DEFAULT_ERROR_PAGE"),
			Category: shared.CategoryCodes,
			Validator: func(code uint) error {
				if code > 999 { //nolint:mnd
					return fmt.Errorf("wrong HTTP code [%d] for the default error page", code)
				}

				return nil
			},
			OnlyOnce: true,
		}
		sendSameHTTPCodeFlag = cli.BoolFlag{
			Name: "send-same-http-code",
			Usage: "The HTTP response should have the same status code as the requested error page (by default, " +
				"every response with an error page will have a status code of 200)",
			Value:    cfg.RespondWithSameHTTPCode,
			Sources:  env("SEND_SAME_HTTP_CODE"),
			Category: shared.CategoryOther,
			OnlyOnce: true,
		}
		showDetailsFlag = cli.BoolFlag{
			Name:     "show-details",
			Usage:    "Show request details in the error page response (if supported by the template)",
			Value:    cfg.ShowDetails,
			Sources:  env("SHOW_DETAILS"),
			Category: shared.CategoryOther,
			OnlyOnce: true,
		}
		proxyHeadersListFlag = cli.StringFlag{
			Name: "proxy-headers",
			Usage: "HTTP headers listed here will be proxied from the original request to the error page response " +
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
			Category: shared.CategoryOther,
			OnlyOnce: true,
			Config:   trim,
		}
		rotationModeFlag = cli.StringFlag{
			Name:     "rotation-mode",
			Value:    config.RotationModeDisabled.String(),
			Usage:    "Templates automatic rotation mode (" + strings.Join(config.RotationModeStrings(), "/") + ")",
			Sources:  env("TEMPLATES_ROTATION_MODE"),
			Category: shared.CategoryTemplates,
			OnlyOnce: true,
			Config:   trim,
			Validator: func(s string) error {
				if _, err := config.ParseRotationMode(s); err != nil {
					return err
				}

				return nil
			},
		}
		readBufferSizeFlag = cli.UintFlag{
			Name: "read-buffer-size",
			Usage: "Per-connection buffer size in bytes for reading requests, this also limits the maximum header size " +
				"(increase this buffer if your clients send multi-KB Request URIs and/or multi-KB headers (e.g., " +
				"large cookies), note that increasing this value will increase memory consumption)",
			Value:    1024 * 5, //nolint:mnd // 5 KB
			Sources:  env("READ_BUFFER_SIZE"),
			Category: shared.CategoryOther,
			OnlyOnce: true,
		}
	)

	// override some flag usage messages
	addrFlag.Usage = "The HTTP server will listen on this IP (v4 or v6) address (set 127.0.0.1/::1 for localhost, " +
		"0.0.0.0 to listen on all interfaces, or specify a custom IP)"
	portFlag.Usage = "The TCP port number for the HTTP server to listen on (0-65535)"

	disableL10nFlag.Value = cfg.L10n.Disable // set the default value depending on the configuration

	cmd.c = &cli.Command{
		Name:    "serve",
		Aliases: []string{"s", "server", "http"},
		Usage:   "Please start the HTTP server to serve the error pages. You can configure various options - please RTFM :D",
		Suggest: true,
		Action: func(ctx context.Context, c *cli.Command) error {
			cmd.opt.http.addr = c.String(addrFlag.Name)
			cmd.opt.http.port = uint16(c.Uint(portFlag.Name)) //nolint:gosec
			cmd.opt.http.readBufferSize = c.Uint(readBufferSizeFlag.Name)
			cfg.L10n.Disable = c.Bool(disableL10nFlag.Name)
			cfg.DefaultCodeToRender = uint16(c.Uint(defaultCodeToRenderFlag.Name)) //nolint:gosec
			cfg.RespondWithSameHTTPCode = c.Bool(sendSameHTTPCodeFlag.Name)
			cfg.RotationMode, _ = config.ParseRotationMode(c.String(rotationModeFlag.Name))
			cfg.ShowDetails = c.Bool(showDetailsFlag.Name)
			cfg.DisableMinification = c.Bool(disableMinificationFlag.Name)

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

				cfg.ProxyHeaders = make([]string, 0, len(m)) // clear the list before adding new headers

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
			&readBufferSizeFlag,
			&disableMinificationFlag,
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	var srv = appHttp.NewServer(log, cmd.opt.http.readBufferSize)

	if err := srv.Register(cfg); err != nil {
		return err
	}

	var startingErrCh = make(chan error, 1) // channel for server starting error
	defer close(startingErrCh)

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
