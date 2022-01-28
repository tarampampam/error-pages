// Package metrics contains custom prometheus metrics and registry factories.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// NewRegistry creates new prometheus registry with pre-registered common collectors.
func NewRegistry() *prometheus.Registry {
	registry := prometheus.NewRegistry()

	// register common metric collectors
	registry.MustRegister(
		// collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return registry
}
