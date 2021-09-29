package http

import (
	"context"
	"strconv"
	"time"

	"github.com/fasthttp/router"
	"github.com/tarampampam/error-pages/internal/checkers"
	"github.com/tarampampam/error-pages/internal/http/common"
	errorpageHandler "github.com/tarampampam/error-pages/internal/http/handlers/errorpage"
	healthzHandler "github.com/tarampampam/error-pages/internal/http/handlers/healthz"
	versionHandler "github.com/tarampampam/error-pages/internal/http/handlers/version"
	"github.com/tarampampam/error-pages/internal/tpl"
	"github.com/tarampampam/error-pages/internal/version"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	fast   *fasthttp.Server
	router *router.Router
}

const (
	defaultWriteTimeout = time.Second * 7
	defaultReadTimeout  = time.Second * 7
	defaultIdleTimeout  = time.Second * 15
)

func NewServer(log *zap.Logger) Server {
	r := router.New()

	return Server{
		// fasthttp docs: <https://github.com/valyala/fasthttp>
		fast: &fasthttp.Server{
			WriteTimeout:          defaultWriteTimeout,
			ReadTimeout:           defaultReadTimeout,
			IdleTimeout:           defaultIdleTimeout,
			Handler:               common.LogRequest(r.Handler, log),
			NoDefaultServerHeader: true,
			ReduceMemoryUsage:     true,
			CloseOnShutdown:       true,
			Logger:                zap.NewStdLog(log),
		},
		router: r,
	}
}

// Start server.
func (s *Server) Start(ip string, port uint16) error {
	return s.fast.ListenAndServe(ip + ":" + strconv.Itoa(int(port)))
}

// Register server routes, middlewares, etc.
// Router docs: <https://github.com/fasthttp/router>
func (s *Server) Register(
	templateName string,
	templates map[string][]byte,
	codes map[string]tpl.Annotator,
) error {
	s.router.GET("/", func(ctx *fasthttp.RequestCtx) {
		common.HandleInternalHTTPError(
			ctx,
			fasthttp.StatusNotFound,
			"Hi there! Error pages are available at the following URLs: /{code}.html",
		)
	})

	s.router.NotFound = func(ctx *fasthttp.RequestCtx) {
		common.HandleInternalHTTPError(
			ctx,
			fasthttp.StatusNotFound,
			"Wrong request URL. Error pages are available at the following URLs: /{code}.html",
		)
	}

	s.router.GET("/version", versionHandler.NewHandler(version.Version()))
	s.router.ANY("/health/live", healthzHandler.NewHandler(checkers.NewLiveChecker()))

	if h, err := errorpageHandler.NewHandler(templateName, templates, codes); err != nil {
		return err
	} else {
		s.router.GET("/{code}.html", h)
	}

	return nil
}

// Stop server.
func (s *Server) Stop(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout) // TODO replace with simple time.After
	defer cancel()

	ch := make(chan error, 1) // channel for server stopping error

	go func() { defer close(ch); ch <- s.fast.Shutdown() }()

	select {
	case err := <-ch:
		return err

	case <-ctx.Done():
		return ctx.Err()
	}
}
