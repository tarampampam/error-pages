package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/tarampampam/error-pages/internal/metrics"
)

func TestNewRegistry(t *testing.T) {
	reg, m := metrics.NewRegistry(), metrics.NewMetrics()

	if err := m.Register(reg); err != nil {
		t.Fatal(err)
	}

	count, err := testutil.GatherAndCount(reg)

	assert.NoError(t, err)
	assert.True(t, count >= 6, "not enough common metrics")
}
