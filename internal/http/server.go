package http

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"gh.tarampamp.am/error-pages/internal/checkers"
	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/http/common"
	errorpageHandler "gh.tarampamp.am/error-pages/internal/http/handlers/errorpage"
	healthzHandler "gh.tarampamp.am/error-pages/internal/http/handlers/healthz"
	indexHandler "gh.tarampamp.am/error-pages/internal/http/handlers/index"
	metricsHandler "gh.tarampamp.am/error-pages/internal/http/handlers/metrics"
	notfoundHandler "gh.tarampamp.am/error-pages/internal/http/handlers/notfound"
	versionHandler "gh.tarampamp.am/error-pages/internal/http/handlers/version"
	"gh.tarampamp.am/error-pages/internal/metrics"
	"gh.tarampamp.am/error-pages/internal/options"
	"gh.tarampamp.am/error-pages/internal/tpl"
	"gh.tarampamp.am/error-pages/internal/version"
)

type Server struct {
	log    *zap.Logger
	fast   *fasthttp.Server
	router *router.Router
	rdr    *tpl.TemplateRenderer
}

const (
	defaultWriteTimeout = time.Second * 4
	defaultReadTimeout  = time.Second * 4
	defaultIdleTimeout  = time.Second * 6
)

func NewServer(log *zap.Logger, readBufferSize int) Server {
	rdr := tpl.NewTemplateRenderer()

	return Server{
		// fasthttp docs: <https://github.com/valyala/fasthttp>
		fast: &fasthttp.Server{
			WriteTimeout:          defaultWriteTimeout,
			ReadBufferSize:        readBufferSize,
			ReadTimeout:           defaultReadTimeout,
			IdleTimeout:           defaultIdleTimeout,
			NoDefaultServerHeader: true,
			ReduceMemoryUsage:     true,
			CloseOnShutdown:       true,
			Logger:                zap.NewStdLog(log),
		},
		router: router.New(),
		log:    log,
		rdr:    rdr,
	}
}

// Start server.
func (s *Server) Start(ip string, port uint16) (err error) {
	if net.ParseIP(ip) == nil {
		return errors.New("invalid IP address")
	}

	var ln net.Listener

	if strings.Count(ip, ":") >= 2 { //nolint:gomnd // ipv6
		if ln, err = net.Listen("tcp6", fmt.Sprintf("[%s]:%d", ip, port)); err != nil {
			return err
		}
	} else { // ipv4
		if ln, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", ip, port)); err != nil {
			return err
		}
	}

	return s.fast.Serve(ln)
}

type templatePicker interface {
	// Pick the template name for responding.
	Pick() string
}

// Register server routes, middlewares, etc.
// Router docs: <https://github.com/fasthttp/router>
func (s *Server) Register(cfg *config.Config, templatePicker templatePicker, opt options.ErrorPage) error {
	reg, m := metrics.NewRegistry(), metrics.NewMetrics()

	if err := m.Register(reg); err != nil {
		return err
	}

	s.fast.Handler = common.DurationMetrics(common.LogRequest(s.router.Handler, s.log), &m)

	s.router.GET("/", indexHandler.NewHandler(cfg, templatePicker, s.rdr, opt))
	s.router.GET("/{code}.html", errorpageHandler.NewHandler(cfg, templatePicker, s.rdr, opt))

	s.router.GET("/version", versionHandler.NewHandler(version.Version()))

	liveHandler := healthzHandler.NewHandler(checkers.NewLiveChecker())
	s.router.ANY("/healthz", liveHandler)
	s.router.ANY("/health/live", liveHandler) // deprecated

	s.router.GET("/metrics", metricsHandler.NewHandler(reg))

	// use index handler to catch all paths? Uses DEFAULT_ERROR_PAGE
	if opt.CatchAll {
		s.router.NotFound = indexHandler.NewHandler(cfg, templatePicker, s.rdr, opt)
	} else {
		// use default not found handler
		s.router.NotFound = notfoundHandler.NewHandler(cfg, templatePicker, s.rdr, opt)
	}

	return nil
}

// Stop server.
func (s *Server) Stop() error {
	if err := s.rdr.Close(); err != nil {
		defer func() { _ = s.fast.Shutdown() }()

		return err
	}

	return s.fast.Shutdown()
}
