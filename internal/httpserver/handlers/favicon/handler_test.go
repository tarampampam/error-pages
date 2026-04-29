package favicon_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/favicon"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	ico, err := os.ReadFile("favicon.ico")
	assert.NoError(t, err)

	icoLen := strconv.Itoa(len(ico))

	h := favicon.New()

	for name, tc := range map[string]struct {
		giveMethod  string
		wantStatus  int
		wantHeaders map[string]string
		wantBody    []byte
	}{
		"GET returns favicon": {
			giveMethod: http.MethodGet,
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":   "image/x-icon",
				"Cache-Control":  "public, max-age=31536000, immutable",
				"Content-Length": icoLen,
			},
			wantBody: ico,
		},
		"HEAD returns headers only": {
			giveMethod: http.MethodHead,
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":   "image/x-icon",
				"Cache-Control":  "public, max-age=31536000, immutable",
				"Content-Length": icoLen,
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

			req := httptest.NewRequest(tc.giveMethod, "/favicon.ico", nil)
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
