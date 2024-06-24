package logger_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestAttrs(t *testing.T) {
	t.Parallel()

	var someTime, _ = time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")

	for name, _tt := range map[string]struct {
		giveAttr logger.Attr

		wantKey   string
		wantValue any
	}{
		"String":   {logger.String("key", "value"), "key", "value"},
		"Strings":  {logger.Strings("key", "value1", "value2"), "key", []string{"value1", "value2"}},
		"Int64":    {logger.Int64("key", 42), "key", int64(42)},
		"Int":      {logger.Int("key", 42), "key", int64(42)},
		"Uint64":   {logger.Uint64("key", 42), "key", uint64(42)},
		"Uint16":   {logger.Uint16("key", 42), "key", uint64(42)},
		"Float64":  {logger.Float64("key", 42.42), "key", 42.42},
		"Bool":     {logger.Bool("key", true), "key", true},
		"Time":     {logger.Time("key", someTime), "key", someTime},
		"Duration": {logger.Duration("key", time.Second), "key", time.Second},
		"Any":      {logger.Any("key", "value"), "key", "value"},
	} {
		tt := _tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantKey, tt.giveAttr.Key)
			assert.Equal(t, tt.wantValue, tt.giveAttr.Value.Any())
		})
	}
}
