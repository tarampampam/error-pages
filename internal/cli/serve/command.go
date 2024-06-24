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
		addrFlag       = shared.ListenAddrFlag
		portFlag       = shared.ListenPortFlag
		addTplFlag     = shared.AddTemplateFlag
		addCodeFlag    = shared.AddHTTPCodeFlag
		jsonFormatFlag = cli.StringFlag{
			Name:     "json-format",
			Usage:    "override the default error page response in JSON format (Go templates are supported)",
			Sources:  env("RESPONSE_JSON_FORMAT"),
			OnlyOnce: true,
			Config:   trim,
		}
		xmlFormatFlag = cli.StringFlag{
			Name:     "xml-format",
			Usage:    "override the default error page response in XML format (Go templates are supported)",
			Sources:  env("RESPONSE_XML_FORMAT"),
			OnlyOnce: true,
			Config:   trim,
		}
		templateNameFlag = cli.StringFlag{
			Name:     "template-name",
			Aliases:  []string{"t"},
			Value:    cfg.TemplateName,
			Usage:    "name of the template to use for rendering error pages",
			Sources:  env("TEMPLATE_NAME"),
			OnlyOnce: true,
			Config:   trim,
		}
		disableL10nFlag = cli.BoolFlag{
			Name:     "disable-l10n",
			Usage:    "disable localization of error pages (if the template supports localization)",
			Value:    cfg.L10n.Disable,
			Sources:  env("DISABLE_L10N"),
			OnlyOnce: true,
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

		// readBufferSizeFlag = cli.UintFlag{
		//	Name: "read-buffer-size",
		//	Usage: "customize the HTTP read buffer size (set per connection for reading requests, also limits the " +
		//		"maximum header size; consider increasing it if your clients send multi-KB request URIs or multi-KB " +
		//		"headers, such as large cookies)",
		//	DefaultText: "not set",
		//	Sources:     cli.EnvVars("READ_BUFFER_SIZE"),
		//	OnlyOnce:    true,
		// }
	)

	cmd.c = &cli.Command{
		Name:    "serve",
		Aliases: []string{"s", "server", "http"},
		Usage:   "Start HTTP server",
		Suggest: true,
		Action: func(ctx context.Context, c *cli.Command) error {
			cmd.opt.http.addr = c.String(addrFlag.Name)
			cmd.opt.http.port = uint16(c.Uint(portFlag.Name))
			// cmd.opt.http.readBufferSize = uint(c.Uint(readBufferSizeFlag.Name))

			cfg.TemplateName = c.String(templateNameFlag.Name)
			cfg.L10n.Disable = c.Bool(disableL10nFlag.Name)
			cfg.DefaultCodeToRender = uint16(c.Uint(defaultCodeToRenderFlag.Name))
			cfg.RespondWithSameHTTPCode = c.Bool(sendSameHTTPCodeFlag.Name)
			cfg.RotationMode, _ = config.ParseRotationMode(c.String(rotationModeFlag.Name))
			cfg.ShowDetails = c.Bool(showDetailsFlag.Name)

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

			if add := c.StringSlice(addTplFlag.Name); len(add) > 0 { // add templates from files to the config
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

			if add := c.StringMap(addCodeFlag.Name); len(add) > 0 { // add custom HTTP codes
				for code, msgAndDesc := range add {
					var (
						parts = strings.SplitN(msgAndDesc, "/", 2) //nolint:mnd
						desc  config.CodeDescription
					)

					if len(parts) > 0 {
						desc.Message = strings.TrimSpace(parts[0])
					}

					if len(parts) > 1 {
						desc.Description = strings.TrimSpace(parts[1])
					}

					cfg.Codes[code] = desc

					log.Info("HTTP code added",
						logger.String("code", code),
						logger.String("message", desc.Message),
						logger.String("description", desc.Description),
					)
				}
			}

			{ // override default JSON and XML formats
				if c.IsSet(jsonFormatFlag.Name) {
					cfg.Formats.JSON = c.String(jsonFormatFlag.Name)
				}

				if c.IsSet(xmlFormatFlag.Name) {
					cfg.Formats.XML = c.String(xmlFormatFlag.Name)
				}
			}

			log.Debug("Configuration",
				logger.Strings("loaded templates", cfg.Templates.Names()...),
				logger.Strings("described HTTP codes", cfg.Codes.Codes()...),
				logger.String("JSON format", cfg.Formats.JSON),
				logger.String("XML format", cfg.Formats.XML),
				logger.String("template name", cfg.TemplateName),
				logger.Bool("disable localization", cfg.L10n.Disable),
				logger.Uint16("default code to render", cfg.DefaultCodeToRender),
				logger.Bool("respond with the same HTTP code", cfg.RespondWithSameHTTPCode),
				logger.Bool("show details", cfg.ShowDetails),
				logger.Strings("proxy HTTP headers", cfg.ProxyHeaders...),
			)

			return cmd.Run(ctx, log, &cfg)
		},
		Flags: []cli.Flag{
			&addrFlag,
			&portFlag,
			&addTplFlag,
			&addCodeFlag,
			&jsonFormatFlag,
			&xmlFormatFlag,
			&templateNameFlag,
			&disableL10nFlag,
			&defaultCodeToRenderFlag,
			&sendSameHTTPCodeFlag,
			&showDetailsFlag,
			&proxyHeadersListFlag,
			&rotationModeFlag,
			// &readBufferSizeFlag,
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	var srv = appHttp.NewServer(ctx, log)

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
