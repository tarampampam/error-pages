package live_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/live"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	okBody := []byte("OK\n")
	okLen := strconv.Itoa(len(okBody))

	h := live.New()

	for name, tc := range map[string]struct {
		giveMethod  string
		wantStatus  int
		wantHeaders map[string]string
		wantBody    []byte
	}{
		"GET returns OK": {
			giveMethod: http.MethodGet,
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":   "text/plain; charset=utf-8",
				"Content-Length": okLen,
			},
			wantBody: okBody,
		},
		"HEAD returns headers only": {
			giveMethod: http.MethodHead,
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":   "text/plain; charset=utf-8",
				"Content-Length": okLen,
			},
			wantBody: []byte{},
		},
		"POST is not allowed": {
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{"Allow": "GET, HEAD"},
			wantBody:    []byte("Method Not Allowed\n"),
		},
		"PUT is not allowed": {
			giveMethod:  http.MethodPut,
			wantStatus:  http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{"Allow": "GET, HEAD"},
			wantBody:    []byte("Method Not Allowed\n"),
		},
		"DELETE is not allowed": {
			giveMethod:  http.MethodDelete,
			wantStatus:  http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{"Allow": "GET, HEAD"},
			wantBody:    []byte("Method Not Allowed\n"),
		},
		"PATCH is not allowed": {
			giveMethod:  http.MethodPatch,
			wantStatus:  http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{"Allow": "GET, HEAD"},
			wantBody:    []byte("Method Not Allowed\n"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.giveMethod, "/healthz", nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			for header, want := range tc.wantHeaders {
				assert.Equal(t, want, rec.Header().Get(header))
			}

			assert.Equal(t, string(tc.wantBody), rec.Body.String())
		})
	}
}
