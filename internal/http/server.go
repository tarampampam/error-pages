package http

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"gh.tarampamp.am/error-pages/internal/appmeta"
	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/http/handlers/error_page"
	"gh.tarampamp.am/error-pages/internal/http/handlers/live"
	"gh.tarampamp.am/error-pages/internal/http/handlers/version"
	"gh.tarampamp.am/error-pages/internal/http/middleware/logreq"
)

// Server is an HTTP server for serving error pages.
type Server struct {
	log    *zap.Logger
	server *http.Server
	mux    *http.ServeMux
}

// NewServer creates a new HTTP server.
func NewServer(baseCtx context.Context, log *zap.Logger) Server {
	const (
		readTimeout    = 30 * time.Second
		writeTimeout   = readTimeout + 10*time.Second // should be bigger than the read timeout
		maxHeaderBytes = (1 << 20) * 5                //nolint:mnd // 5 MB
	)

	var (
		mux = http.NewServeMux()
		srv = &http.Server{
			Handler:           mux,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			ReadHeaderTimeout: readTimeout,
			MaxHeaderBytes:    maxHeaderBytes,
			ErrorLog:          zap.NewStdLog(log),
			BaseContext:       func(net.Listener) context.Context { return baseCtx },
		}
	)

	return Server{log: log, server: srv, mux: mux}
}

// Register server handlers, middlewares, etc.
func (s *Server) Register(cfg *config.Config) error {
	// register middleware
	s.server.Handler = logreq.New(s.log, func(r *http.Request) bool {
		// skip logging healthcheck requests
		return strings.Contains(strings.ToLower(r.UserAgent()), "healthcheck")
	})(s.server.Handler)

	{ // register handlers (https://go.dev/blog/routing-enhancements)
		var errorPageHandler = error_page.New()

		s.mux.Handle("/", errorPageHandler)
		s.mux.Handle("/{any}", errorPageHandler)

		var liveHandler = live.New()

		s.mux.Handle("GET /health/live", liveHandler)
		s.mux.Handle("GET /healthz", liveHandler)
		s.mux.Handle("GET /live", liveHandler)

		s.mux.Handle("GET /version", version.New(appmeta.Version()))
	}

	return nil
}

// Start server.
func (s *Server) Start(ip string, port uint16) error {
	s.server.Addr = ip + ":" + strconv.Itoa(int(port))

	return s.server.ListenAndServe()
}

// Stop server gracefully.
func (s *Server) Stop(timeout time.Duration) error {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
