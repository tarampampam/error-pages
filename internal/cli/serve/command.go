package serve

import (
	"context"
	"errors"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/tarampampam/error-pages/internal/breaker"
	"github.com/tarampampam/error-pages/internal/config"
	appHttp "github.com/tarampampam/error-pages/internal/http"
	"github.com/tarampampam/error-pages/internal/pick"
	"github.com/tarampampam/error-pages/internal/tpl"
	"go.uber.org/zap"
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
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if configFile == nil {
				return errors.New("path to the config file is required for this command")
			}

			if err := f.overrideUsingEnv(cmd.Flags()); err != nil {
				return err
			}

			if c, err := config.FromYamlFile(*configFile); err != nil {
				return err
			} else {
				if err = c.Validate(); err != nil {
					return err
				}

				cfg = c
			}

			return f.validate()
		},
		RunE: func(*cobra.Command, []string) error { return run(ctx, log, f, cfg) },
	}

	f.init(cmd.Flags())

	return cmd
}

// run current command.
func run(parentCtx context.Context, log *zap.Logger, f flags, cfg *config.Config) error { //nolint:funlen
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
		errorPages    = tpl.NewErrorPages()
		templateNames = make([]string, 0) // slice with all possible template names
	)

	log.Debug("Loading templates")

	if templates, err := cfg.LoadTemplates(); err == nil {
		if len(templates) > 0 {
			for templateName, content := range templates {
				errorPages.AddTemplate(templateName, content)
				templateNames = append(templateNames, templateName)
			}

			for code, desc := range cfg.Pages {
				errorPages.AddPage(code, desc.Message, desc.Description)
			}

			log.Info("Templates loaded", zap.Int("templates", len(templates)), zap.Int("pages", len(cfg.Pages)))
		} else {
			return errors.New("no loaded templates")
		}
	} else {
		return err
	}

	sort.Strings(templateNames) // sorting is important for the first template picking

	var picker *pick.StringsSlice

	switch f.template.name {
	case useRandomTemplate:
		log.Info("A random template will be used")

		picker = pick.NewStringsSlice(templateNames, pick.RandomOnce)

	case useRandomTemplateOnEachRequest:
		log.Info("A random template on EACH request will be used")

		picker = pick.NewStringsSlice(templateNames, pick.RandomEveryTime)

	case "":
		log.Info("The first template (ordered by name) will be used")

		picker = pick.NewStringsSlice(templateNames, pick.First)

	default:
		var found bool

		for i := 0; i < len(templateNames); i++ {
			if templateNames[i] == f.template.name {
				found = true

				break
			}
		}

		if !found {
			return errors.New("requested nonexistent template: " + f.template.name)
		}

		log.Info("We will use the requested template", zap.String("name", f.template.name))
		picker = pick.NewStringsSlice([]string{f.template.name}, pick.First)
	}

	// create HTTP server
	server := appHttp.NewServer(log)

	// register server routes, middlewares, etc.
	server.Register(&errorPages, picker, f.defaultErrorPage)

	startedAt, startingErrCh := time.Now(), make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", f.listen.ip),
			zap.Uint16("port", f.listen.port),
			zap.String("default error page", f.defaultErrorPage),
		)

		if err := server.Start(f.listen.ip, f.listen.port); err != nil {
			errCh <- err
		}
	}(startingErrCh)

	// and wait for...
	select {
	case err := <-startingErrCh: // ..server starting error
		return err

	case <-ctx.Done(): // ..or context cancellation
		log.Info("Gracefully server stopping", zap.Duration("uptime", time.Since(startedAt)))

		// stop the server using created context above
		if err := server.Stop(); err != nil {
			return err
		}
	}

	return nil
}
