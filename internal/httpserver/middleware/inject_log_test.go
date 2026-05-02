package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/httpserver/middleware"
	"gh.tarampamp.am/error-pages/v4/internal/logger"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNewInjectLog(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveLogger        *logger.Logger
		giveSecondLogger  *logger.Logger
		giveDoubleWrapped bool
		wantLoggerSet     bool
	}{
		"injects logger into context": {
			giveLogger:    logger.NewNop(),
			wantLoggerSet: true,
		},
		"first logger wins when middleware is applied twice": {
			giveLogger:        logger.NewNop(),
			giveSecondLogger:  logger.NewNop(),
			giveDoubleWrapped: true,
			wantLoggerSet:     true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// wantEffectiveLogger is always the outer (first) logger; set here to keep
			// test case declarations free of redundant field assignments.
			wantEffectiveLogger := tc.giveLogger

			var (
				capturedLoggerSet bool
				capturedLogger    *logger.Logger
			)

			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				capturedLoggerSet = logger.IsSet(r.Context())
				capturedLogger = logger.FromContext(r.Context())
			})

			var handler http.Handler
			if tc.giveDoubleWrapped {
				// wrap next with the inner middleware first, then the outer one on top -
				// the outer middleware runs first and should win.
				handler = middleware.NewInjectLog(tc.giveLogger)(
					middleware.NewInjectLog(tc.giveSecondLogger)(next),
				)
			} else {
				handler = middleware.NewInjectLog(tc.giveLogger)(next)
			}

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", http.NoBody))

			assert.Equal(t, tc.wantLoggerSet, capturedLoggerSet)

			if tc.wantLoggerSet {
				// pointer identity: the exact logger injected by the outer middleware must be the one in context
				assert.Equal(t, wantEffectiveLogger, capturedLogger)
			}
		})
	}
}
