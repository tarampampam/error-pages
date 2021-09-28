package serve

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/tarampampam/error-pages/internal/http/handlers/errorpage"
	"github.com/tarampampam/error-pages/internal/tpl"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/tarampampam/error-pages/internal/breaker"
	"github.com/tarampampam/error-pages/internal/config"
	appHttp "github.com/tarampampam/error-pages/internal/http"
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

const serverShutdownTimeout = 15 * time.Second

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

	// load templates content
	templates, loadingErr := cfg.LoadTemplates()
	if loadingErr != nil {
		return loadingErr
	} else if len(templates) == 0 {
		return errors.New("no loaded templates")
	}

	if f.template.name != "" && f.template.name != errorpage.UseRandom && f.template.name != errorpage.UseRandomOnEachRequest { //nolint:lll
		if _, found := templates[f.template.name]; !found {
			return errors.New("requested nonexistent template: " + f.template.name) // requested unknown template
		}
	}

	// burn the error codes map
	codes := make(map[string]tpl.Annotator)
	for code, desc := range cfg.Pages {
		codes[code] = tpl.Annotator{Message: desc.Message, Description: desc.Description}
	}

	// create HTTP server
	server := appHttp.NewServer(log)

	// register server routes, middlewares, etc.
	if err := server.Register(f.template.name, templates, codes); err != nil {
		return err
	}

	startingErrCh := make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", f.listen.ip),
			zap.Uint16("port", f.listen.port),
			zap.String("template name", f.template.name),
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
		log.Info("Gracefully server stopping")

		stoppedAt := time.Now()

		// stop the server using created context above
		if err := server.Stop(serverShutdownTimeout); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error("Server stopping timeout exceeded", zap.Duration("timeout", serverShutdownTimeout))
			}

			return err
		}

		log.Debug("Server stopped", zap.Duration("stopping duration", time.Since(stoppedAt)))
	}

	return nil
}
