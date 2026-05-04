package httpserver

import (
	"net/http"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/error_page"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/favicon"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/live"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/version"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver/middleware"
	"gh.tarampamp.am/error-pages/v4/internal/logger"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
)

// NewHandler creates a new HTTP handler that serves all server endpoints. It does not use MUX because the
// number of endpoints is small, and the goal is to achieve maximum performance.
func NewHandler(
	log *logger.Logger,
	defaultCode uint16,
	respondSameStatus bool,
	proxyHeaders []string,
	describer error_page.CodeDescriber,
	templater error_page.Templater,
	showDetails bool,
	l10nDisabled bool,
	homepageURL string,
	links []tpl.Link,
) http.Handler {
	const (
		healthzEndpoint    = "/healthz"
		healthEndpoint     = "/health"
		healthLiveEndpoint = "/health/live"
		liveEndpoint       = "/live"

		versionEndpoint = "/version"
		faviconEndpoint = "/favicon.ico"
	)

	liveHandler := live.New()
	versionHandler := version.New(appmeta.Version())
	faviconHandler := favicon.New()
	errorPagesHandler := error_page.New(
		log,
		defaultCode,
		respondSameStatus,
		proxyHeaders,
		describer,
		templater,
		showDetails,
		l10nDisabled,
		homepageURL,
		links,
	)

	return middleware.Apply(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case healthzEndpoint, healthEndpoint, healthLiveEndpoint, liveEndpoint:
				liveHandler.ServeHTTP(w, r)

				return

			case versionEndpoint:
				versionHandler.ServeHTTP(w, r)

				return

			case faviconEndpoint:
				faviconHandler.ServeHTTP(w, r)

				return
			}

			// catch-all handler for error pages
			errorPagesHandler.ServeHTTP(w, r)
		}),
		middleware.NewInjectLog(log),
		middleware.NewAccessLog(logger.InfoLevel, func(r *http.Request) bool {
			// skip logging for the healthz endpoint
			return r.URL.Path == healthzEndpoint ||
				r.URL.Path == healthEndpoint ||
				r.URL.Path == healthLiveEndpoint ||
				r.URL.Path == liveEndpoint
		}),
	)
}
