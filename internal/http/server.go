package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"gh.tarampamp.am/error-pages/internal/appmeta"
	"gh.tarampamp.am/error-pages/internal/config"
	ep "gh.tarampamp.am/error-pages/internal/http/handlers/error_page"
	"gh.tarampamp.am/error-pages/internal/http/handlers/live"
	"gh.tarampamp.am/error-pages/internal/http/handlers/static"
	"gh.tarampamp.am/error-pages/internal/http/handlers/version"
	"gh.tarampamp.am/error-pages/internal/http/middleware/logreq"
	"gh.tarampamp.am/error-pages/internal/logger"
)

// Server is an HTTP server for serving error pages.
type Server struct {
	log        *logger.Logger
	server     *fasthttp.Server
	beforeStop func()
}

// NewServer creates a new HTTP server.
func NewServer(log *logger.Logger, readBufferSize uint) Server {
	const (
		readTimeout  = 30 * time.Second
		writeTimeout = readTimeout + 10*time.Second // should be bigger than the read timeout
	)

	return Server{
		log: log,
		server: &fasthttp.Server{
			ReadTimeout:                  readTimeout,
			WriteTimeout:                 writeTimeout,
			ReadBufferSize:               int(readBufferSize), //nolint:gosec
			DisablePreParseMultipartForm: true,
			NoDefaultServerHeader:        true,
			CloseOnShutdown:              true,
			Logger:                       logger.NewStdLog(log),
		},
		beforeStop: func() {}, // noop
	}
}

// Register server handlers, middlewares, etc.
func (s *Server) Register(cfg *config.Config) error {
	var (
		liveHandler    = live.New()
		versionHandler = version.New(appmeta.Version())
		faviconHandler = static.New(static.Favicon)

		errorPagesHandler, closeCache = ep.New(cfg, s.log)

		notFound   = http.StatusText(http.StatusNotFound) + "\n"
		notAllowed = http.StatusText(http.StatusMethodNotAllowed) + "\n"
	)

	// wrap the before shutdown function to close the cache
	s.beforeStop = closeCache

	s.server.Handler = func(ctx *fasthttp.RequestCtx) {
		var url, method = string(ctx.Path()), string(ctx.Method())

		switch {
		// live endpoints
		case url == "/healthz" || url == "/health/live" || url == "/health" || url == "/live":
			liveHandler(ctx)

		// version endpoint
		case url == "/version":
			versionHandler(ctx)

		// favicon.ico endpoint
		case url == "/favicon.ico":
			faviconHandler(ctx)

		// error pages endpoints:
		//	- /
		//	-	/{code}.html
		//	- /{code}.htm
		//	- /{code}
		//
		// the HTTP method is not limited to GET and HEAD - it can be any
		case url == "/" || ep.URLContainsCode(url) || ep.HeadersContainCode(&ctx.Request.Header):
			errorPagesHandler(ctx)

		// wrong requests handling
		default:
			switch method {
			case fasthttp.MethodHead:
				ctx.Error(notAllowed, fasthttp.StatusNotFound)
			case fasthttp.MethodGet:
				ctx.Error(notFound, fasthttp.StatusNotFound)
			default:
				ctx.Error(notAllowed, fasthttp.StatusMethodNotAllowed)
			}
		}
	}

	// apply middleware
	s.server.Handler = logreq.New(s.log, func(ctx *fasthttp.RequestCtx) bool {
		// skip logging healthcheck and .ico (favicon) requests
		return strings.Contains(strings.ToLower(string(ctx.UserAgent())), "healthcheck") ||
			strings.HasSuffix(string(ctx.Path()), ".ico")
	})(s.server.Handler)

	return nil
}

// Start server.
func (s *Server) Start(ip string, port uint16) (err error) {
	if net.ParseIP(ip) == nil {
		return errors.New("invalid IP address")
	}

	var ln net.Listener

	if strings.Count(ip, ":") >= 2 { //nolint:mnd // ipv6
		if ln, err = net.Listen("tcp6", fmt.Sprintf("[%s]:%d", ip, port)); err != nil {
			return err
		}
	} else { // ipv4
		if ln, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", ip, port)); err != nil {
			return err
		}
	}

	return s.server.Serve(ln)
}

// Stop server gracefully.
func (s *Server) Stop(timeout time.Duration) error {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s.beforeStop()

	return s.server.ShutdownWithContext(ctx)
}
