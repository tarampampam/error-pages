package http

import (
	"strconv"
	"time"

	"github.com/fasthttp/router"
	"github.com/tarampampam/error-pages/internal/checkers"
	"github.com/tarampampam/error-pages/internal/http/common"
	errorpageHandler "github.com/tarampampam/error-pages/internal/http/handlers/errorpage"
	healthzHandler "github.com/tarampampam/error-pages/internal/http/handlers/healthz"
	indexHandler "github.com/tarampampam/error-pages/internal/http/handlers/index"
	notfoundHandler "github.com/tarampampam/error-pages/internal/http/handlers/notfound"
	versionHandler "github.com/tarampampam/error-pages/internal/http/handlers/version"
	"github.com/tarampampam/error-pages/internal/version"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	fast   *fasthttp.Server
	router *router.Router
}

const (
	defaultWriteTimeout = time.Second * 4
	defaultReadTimeout  = time.Second * 4
	defaultIdleTimeout  = time.Second * 6
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

type (
	errorsPager interface {
		// GetPage with passed template name and error code.
		GetPage(templateName, code string) ([]byte, error)
	}

	templatePicker interface {
		// Pick the template name for responding.
		Pick() string
	}
)

// Register server routes, middlewares, etc.
// Router docs: <https://github.com/fasthttp/router>
func (s *Server) Register(
	errorsPager errorsPager,
	templatePicker templatePicker,
	defaultPageCode string,
	defaultHTTPCode uint16,
) {
	s.router.GET("/", indexHandler.NewHandler(errorsPager, templatePicker, defaultPageCode, defaultHTTPCode))
	s.router.GET("/version", versionHandler.NewHandler(version.Version()))
	s.router.ANY("/health/live", healthzHandler.NewHandler(checkers.NewLiveChecker()))
	s.router.GET("/{code}.html", errorpageHandler.NewHandler(errorsPager, templatePicker))

	s.router.NotFound = notfoundHandler.NewHandler()
}

// Stop server.
func (s *Server) Stop() error { return s.fast.Shutdown() }
