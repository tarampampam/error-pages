// Package metrics contains HTTP handler for application metrics (prometheus format) generation.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// NewHandler creates metrics handler.
func NewHandler(registry prometheus.Gatherer) fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}),
	)
}
