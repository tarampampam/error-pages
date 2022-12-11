package serve

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/tarampampam/error-pages/internal/breaker"
	"github.com/tarampampam/error-pages/internal/config"
	appHttp "github.com/tarampampam/error-pages/internal/http"
	"github.com/tarampampam/error-pages/internal/pick"
)

// NewCommand creates `serve` command.
func NewCommand(ctx context.Context, log *zap.Logger, configFile *string) *cobra.Command {
	var (
		f   flags
		cfg *config.Config
	)

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s", "server"},
		Short:   "Start HTTP server",
		PreRunE: func(cmd *cobra.Command, _ []string) (err error) {
			if configFile == nil {
				return errors.New("path to the config file is required for this command")
			}

			if err = f.OverrideUsingEnv(cmd.Flags()); err != nil {
				return err
			}

			if cfg, err = config.FromYamlFile(*configFile); err != nil {
				return err
			}

			return f.Validate()
		},
		RunE: func(*cobra.Command, []string) error { return run(ctx, log, cfg, f) },
	}

	f.Init(cmd.Flags())

	return cmd
}

// run current command.
func run(parentCtx context.Context, log *zap.Logger, cfg *config.Config, f flags) error { //nolint:funlen
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

		opt = f.ToOptions()
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
	server := appHttp.NewServer(log)

	// register server routes, middlewares, etc.
	if err := server.Register(cfg, picker, opt); err != nil {
		return err
	}

	startedAt, startingErrCh := time.Now(), make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", f.Listen.IP),
			zap.Uint16("port", f.Listen.Port),
			zap.String("default error page", opt.Default.PageCode),
			zap.Uint16("default HTTP response code", opt.Default.HTTPCode),
			zap.Strings("proxy headers", opt.ProxyHTTPHeaders),
			zap.Bool("show request details", opt.ShowDetails),
			zap.Bool("catch all pages", opt.CatchAll),
			zap.Bool("localization disabled", opt.L10n.Disabled),
		)

		if err := server.Start(f.Listen.IP, f.Listen.Port); err != nil {
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
