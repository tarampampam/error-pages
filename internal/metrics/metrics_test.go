package metrics_test

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/metrics"
)

func TestMetrics_Register(t *testing.T) {
	var (
		registry = prometheus.NewRegistry()
		m        = metrics.NewMetrics()
	)

	assert.NoError(t, m.Register(registry))

	count, err := testutil.GatherAndCount(registry,
		"http_requests_total_count",
		"http_requests_duration_milliseconds",
	)
	assert.NoError(t, err)

	assert.Equal(t, 2, count)
}

func TestMetrics_IncrementTotalRequests(t *testing.T) {
	p := metrics.NewMetrics()

	p.IncrementTotalRequests()

	metric := getMetric(t, &p, "http_requests_total_count")
	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

func TestMetrics_ObserveRequestDuration(t *testing.T) {
	p := metrics.NewMetrics()

	p.ObserveRequestDuration(time.Second)

	metric := getMetric(t, &p, "http_requests_duration_milliseconds")
	assert.Equal(t, float64(1), metric.Histogram.GetSampleSum())
}

type registerer interface {
	Register(prometheus.Registerer) error
}

func getMetric(t *testing.T, reg registerer, name string) *dto.Metric {
	t.Helper()

	registry := prometheus.NewRegistry()
	_ = reg.Register(registry)

	families, _ := registry.Gather()

	for _, family := range families {
		if family.GetName() == name {
			return family.Metric[0]
		}
	}

	assert.FailNowf(t, "cannot resolve metric for: %s", name)

	return nil
}
