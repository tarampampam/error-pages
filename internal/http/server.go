package http

import (
	"strconv"
	"time"

	"github.com/fasthttp/router"
	"github.com/tarampampam/error-pages/internal/checkers"
	"github.com/tarampampam/error-pages/internal/config"
	"github.com/tarampampam/error-pages/internal/http/common"
	errorpageHandler "github.com/tarampampam/error-pages/internal/http/handlers/errorpage"
	healthzHandler "github.com/tarampampam/error-pages/internal/http/handlers/healthz"
	indexHandler "github.com/tarampampam/error-pages/internal/http/handlers/index"
	metricsHandler "github.com/tarampampam/error-pages/internal/http/handlers/metrics"
	notfoundHandler "github.com/tarampampam/error-pages/internal/http/handlers/notfound"
	versionHandler "github.com/tarampampam/error-pages/internal/http/handlers/version"
	"github.com/tarampampam/error-pages/internal/metrics"
	"github.com/tarampampam/error-pages/internal/version"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	log    *zap.Logger
	fast   *fasthttp.Server
	router *router.Router
}

const (
	defaultWriteTimeout = time.Second * 4
	defaultReadTimeout  = time.Second * 4
	defaultIdleTimeout  = time.Second * 6
)

func NewServer(log *zap.Logger) Server {
	return Server{
		// fasthttp docs: <https://github.com/valyala/fasthttp>
		fast: &fasthttp.Server{
			WriteTimeout:          defaultWriteTimeout,
			ReadTimeout:           defaultReadTimeout,
			IdleTimeout:           defaultIdleTimeout,
			NoDefaultServerHeader: true,
			ReduceMemoryUsage:     true,
			CloseOnShutdown:       true,
			Logger:                zap.NewStdLog(log),
		},
		router: router.New(),
		log:    log,
	}
}

// Start server.
func (s *Server) Start(ip string, port uint16) error {
	return s.fast.ListenAndServe(ip + ":" + strconv.Itoa(int(port)))
}

type templatePicker interface {
	// Pick the template name for responding.
	Pick() string
}

// Register server routes, middlewares, etc.
// Router docs: <https://github.com/fasthttp/router>
func (s *Server) Register(
	cfg *config.Config,
	templatePicker templatePicker,
	defaultPageCode string,
	defaultHTTPCode uint16,
	showDetails bool,
) error {
	reg, m := metrics.NewRegistry(), metrics.NewMetrics()

	if err := m.Register(reg); err != nil {
		return err
	}

	s.fast.Handler = common.DurationMetrics(common.LogRequest(s.router.Handler, s.log), &m)

	s.router.GET("/", indexHandler.NewHandler(cfg, templatePicker, defaultPageCode, defaultHTTPCode, showDetails))
	s.router.GET("/{code}.html", errorpageHandler.NewHandler(cfg, templatePicker, showDetails))
	s.router.GET("/version", versionHandler.NewHandler(version.Version()))

	liveHandler := healthzHandler.NewHandler(checkers.NewLiveChecker())
	s.router.ANY("/healthz", liveHandler)
	s.router.ANY("/health/live", liveHandler) // deprecated

	s.router.GET("/metrics", metricsHandler.NewHandler(reg))

	s.router.NotFound = notfoundHandler.NewHandler()

	return nil
}

// Stop server.
func (s *Server) Stop() error { return s.fast.Shutdown() }
