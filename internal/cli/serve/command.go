package serve

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"gh.tarampamp.am/error-pages/internal/breaker"
	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/env"
	appHttp "gh.tarampamp.am/error-pages/internal/http"
	"gh.tarampamp.am/error-pages/internal/options"
	"gh.tarampamp.am/error-pages/internal/pick"
)

type command struct {
	c *cli.Command
}

const (
	templateNameFlagName     = "template-name"
	defaultErrorPageFlagName = "default-error-page"
	defaultHTTPCodeFlagName  = "default-http-code"
	showDetailsFlagName      = "show-details"
	proxyHTTPHeadersFlagName = "proxy-headers"
	disableL10nFlagName      = "disable-l10n"
	catchAllFlagName         = "catch-all"
	readBufferSizeFlagName   = "read-buffer"
)

const (
	useRandomTemplate              = "random"
	useRandomTemplateOnEachRequest = "i-said-random"
	useRandomTemplateDaily         = "random-daily"
	useRandomTemplateHourly        = "random-hourly"
)

// NewCommand creates `serve` command.
func NewCommand(log *zap.Logger) *cli.Command { //nolint:funlen
	var cmd = command{}

	cmd.c = &cli.Command{
		Name:    "serve",
		Aliases: []string{"s", "server"},
		Usage:   "Start HTTP server",
		Action: func(c *cli.Context) error {
			var cfg *config.Config

			if configPath := c.String(shared.ConfigFileFlag.Name); configPath == "" { // load config from file
				return errors.New("path to the config file is required for this command")
			} else if loadedCfg, err := config.FromYamlFile(c.String(shared.ConfigFileFlag.Name)); err != nil {
				return err
			} else {
				cfg = loadedCfg
			}

			var (
				ip   = c.String(shared.ListenAddrFlag.Name)
				port = uint16(c.Uint(shared.ListenPortFlag.Name))
				readBufferSize = int(c.Int(shared.ReadBufferSizeFlag.Name))
				o    options.ErrorPage
			)

			if net.ParseIP(ip) == nil {
				return fmt.Errorf("wrong IP address [%s] for listening", ip)
			}

			{ // fill options
				o.Template.Name = c.String(templateNameFlagName)
				o.L10n.Disabled = c.Bool(disableL10nFlagName)
				o.Default.PageCode = c.String(defaultErrorPageFlagName)
				o.Default.HTTPCode = uint16(c.Uint(defaultHTTPCodeFlagName))
				o.ShowDetails = c.Bool(showDetailsFlagName)
				o.CatchAll = c.Bool(catchAllFlagName)

				if headers := c.String(proxyHTTPHeadersFlagName); headers != "" { //nolint:nestif
					var m = make(map[string]struct{})

					// make unique and ignore empty strings
					for _, header := range strings.Split(headers, ",") {
						if h := strings.TrimSpace(header); h != "" {
							if strings.ContainsRune(h, ' ') {
								return fmt.Errorf("whitespaces in the HTTP headers for proxying [%s] are not allowed", header)
							}

							if _, ok := m[h]; !ok {
								m[h] = struct{}{}
							}
						}
					}

					// convert map into slice
					o.ProxyHTTPHeaders = make([]string, 0, len(m))
					for h := range m {
						o.ProxyHTTPHeaders = append(o.ProxyHTTPHeaders, h)
					}
				}
			}

			if o.Default.HTTPCode > 599 { //nolint:gomnd
				return fmt.Errorf("wrong default HTTP response code [%d]", o.Default.HTTPCode)
			}

			return cmd.Run(c.Context, log, cfg, ip, port, readBufferSize, o)
		},
		Flags: []cli.Flag{
			shared.ConfigFileFlag,
			shared.ListenPortFlag,
			shared.ListenAddrFlag,
			&cli.StringFlag{
				Name:    templateNameFlagName,
				Aliases: []string{"t"},
				Usage: fmt.Sprintf(
					"template name (set \"%s\" to use a randomized or \"%s\" to use a randomized template on "+
						"each request or \"%s/%s\" daily/hourly randomized)",
					useRandomTemplate,
					useRandomTemplateOnEachRequest,
					useRandomTemplateDaily,
					useRandomTemplateHourly,
				),
				EnvVars: []string{env.TemplateName.String()},
			},
			&cli.StringFlag{
				Name:    defaultErrorPageFlagName,
				Value:   "404",
				Usage:   "default error page",
				EnvVars: []string{env.DefaultErrorPage.String()},
			},
			&cli.UintFlag{
				Name:    defaultHTTPCodeFlagName,
				Value:   404, //nolint:gomnd
				Usage:   "default HTTP response code",
				EnvVars: []string{env.DefaultHTTPCode.String()},
			},
			&cli.BoolFlag{
				Name:    showDetailsFlagName,
				Usage:   "show request details in response",
				EnvVars: []string{env.ShowDetails.String()},
			},
			&cli.StringFlag{
				Name:    proxyHTTPHeadersFlagName,
				Usage:   "proxy HTTP request headers list (comma-separated)",
				EnvVars: []string{env.ProxyHTTPHeaders.String()},
			},
			&cli.BoolFlag{
				Name:    disableL10nFlagName,
				Usage:   "disable error pages localization",
				EnvVars: []string{env.DisableL10n.String()},
			},
			&cli.BoolFlag{
				Name:    catchAllFlagName,
				Usage:   "catch all pages",
				EnvVars: []string{env.CatchAll.String()},
			},
			shared.ReadBufferSizeFlag,
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run( //nolint:funlen
	parentCtx context.Context, log *zap.Logger, cfg *config.Config, ip string, port uint16, readBufferSize int, opt options.ErrorPage,
) error {
	var (
		ctx, cancel = context.WithCancel(parentCtx) // serve context creation
		oss         = breaker.NewOSSignals(ctx)     // OS signals listener
	)

	// subscribe for system signals
	oss.Subscribe(func(sig os.Signal) {
		log.Warn("Stopping by OS signal..", zap.String("signal", sig.String()))

		cancel()
	})

	defer func() {
		cancel()   // call the cancellation function after all
		oss.Stop() // stop system signals listening
	}()

	var (
		templateNames = cfg.TemplateNames()
		picker        interface{ Pick() string }
	)

	switch opt.Template.Name {
	case useRandomTemplate:
		log.Info("A random template will be used")

		picker = pick.NewStringsSlice(templateNames, pick.RandomOnce)

	case useRandomTemplateOnEachRequest:
		log.Info("A random template on EACH request will be used")

		picker = pick.NewStringsSlice(templateNames, pick.RandomEveryTime)

	case useRandomTemplateDaily:
		log.Info("A random template will be used and changed once a day")

		picker = pick.NewStringsSliceWithInterval(templateNames, pick.RandomEveryTime, time.Hour*24) //nolint:gomnd

	case useRandomTemplateHourly:
		log.Info("A random template will be used and changed hourly")

		picker = pick.NewStringsSliceWithInterval(templateNames, pick.RandomEveryTime, time.Hour)

	case "":
		log.Info("The first template (ordered by name) will be used")

		picker = pick.NewStringsSlice(templateNames, pick.First)

	default:
		if t, found := cfg.Template(opt.Template.Name); found {
			log.Info("We will use the requested template", zap.String("name", t.Name()))
			picker = pick.NewStringsSlice([]string{t.Name()}, pick.First)
		} else {
			return errors.New("requested nonexistent template: " + opt.Template.Name)
		}
	}

	// create HTTP server
	server := appHttp.NewServer(log, readBufferSize)

	// register server routes, middlewares, etc.
	if err := server.Register(cfg, picker, opt); err != nil {
		return err
	}

	startedAt, startingErrCh := time.Now(), make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", ip),
			zap.Uint16("port", port),
			zap.String("default error page", opt.Default.PageCode),
			zap.Uint16("default HTTP response code", opt.Default.HTTPCode),
			zap.Strings("proxy headers", opt.ProxyHTTPHeaders),
			zap.Bool("show request details", opt.ShowDetails),
			zap.Bool("localization disabled", opt.L10n.Disabled),
			zap.Bool("catch all enabled", opt.CatchAll),
			zap.Int("read buffer size", readBufferSize),
		)

		if err := server.Start(ip, port); err != nil {
			errCh <- err
		}
	}(startingErrCh)

	// and wait for...
	select {
	case err := <-startingErrCh: // ..server starting error
		return err

	case <-ctx.Done(): // ..or context cancellation
		log.Info("Gracefully server stopping", zap.Duration("uptime", time.Since(startedAt)))

		if p, ok := picker.(interface{ Close() error }); ok {
			if err := p.Close(); err != nil {
				return err
			}
		}

		// stop the server using created context above
		if err := server.Stop(); err != nil {
			return err
		}
	}

	return nil
}
