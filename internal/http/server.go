package http

import (
	"context"
	"net"
	"net/http"
	"regexp"
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
}

// NewServer creates a new HTTP server.
func NewServer(baseCtx context.Context, log *zap.Logger) Server {
	const (
		readTimeout    = 30 * time.Second
		writeTimeout   = readTimeout + 10*time.Second // should be bigger than the read timeout
		maxHeaderBytes = (1 << 20) * 5                //nolint:mnd // 5 MB
	)

	return Server{
		log: log,
		server: &http.Server{
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			ReadHeaderTimeout: readTimeout,
			MaxHeaderBytes:    maxHeaderBytes,
			ErrorLog:          zap.NewStdLog(log),
			BaseContext:       func(net.Listener) context.Context { return baseCtx },
		},
	}
}

// Register server handlers, middlewares, etc.
func (s *Server) Register(cfg *config.Config) error {
	var (
		liveHandler       = live.New()
		versionHandler    = version.New(appmeta.Version())
		errorPagesHandler = error_page.New()

		errorPageRegex = regexp.MustCompile(`^/(\d{3})(?:\.html|\.htm)?$`) // TODO: rewrite to function
	)

	s.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var url, method = r.URL.Path, r.Method

		switch {
		// live endpoints
		case url == "/health/live" || url == "/health" || url == "/healthz" || url == "/live":
			liveHandler.ServeHTTP(w, r)
		// version endpoint
		case url == "/version":
			versionHandler.ServeHTTP(w, r)
		// error pages endpoints:
		//	- /
		//	-	/{code}.html
		//	- /{code}.htm
		//	- /{code}
		case method == http.MethodGet && (url == "/" || errorPageRegex.MatchString(url)):
			errorPagesHandler.ServeHTTP(w, r)
		// wrong requests handling
		default:
			switch {
			case method == http.MethodHead:
				w.WriteHeader(http.StatusNotFound)
			case method == http.MethodGet:
				http.NotFound(w, r)
			default:
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			}
		}
	})

	// apply middleware
	s.server.Handler = logreq.New(s.log, func(r *http.Request) bool {
		// skip logging healthcheck requests
		return strings.Contains(strings.ToLower(r.UserAgent()), "healthcheck")
	})(s.server.Handler)

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
