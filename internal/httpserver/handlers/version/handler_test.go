package version_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/version"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	const ver = "1.2.3"

	verBody := []byte(`{"version":"` + ver + `"}`)
	verLen := strconv.Itoa(len(verBody))

	h := version.New(ver)

	for name, tc := range map[string]struct {
		giveMethod  string
		wantStatus  int
		wantHeaders map[string]string
		wantBody    []byte
	}{
		"GET returns version JSON": {
			giveMethod: http.MethodGet,
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":   "application/json; charset=utf-8",
				"Content-Length": verLen,
			},
			wantBody: verBody,
		},
		"HEAD returns headers only": {
			giveMethod: http.MethodHead,
			wantStatus: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type":   "application/json; charset=utf-8",
				"Content-Length": verLen,
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

			req := httptest.NewRequest(tc.giveMethod, "/version", nil)
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

func TestNew_DifferentVersions(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveVersion string
		wantBody    []byte
	}{
		"empty version": {
			giveVersion: "",
			wantBody:    []byte(`{"version":""}`),
		},
		"semver": {
			giveVersion: "2.0.0",
			wantBody:    []byte(`{"version":"2.0.0"}`),
		},
		"dev build": {
			giveVersion: "dev-abc123",
			wantBody:    []byte(`{"version":"dev-abc123"}`),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/version", nil)
			rec := httptest.NewRecorder()

			version.New(tc.giveVersion).ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, strconv.Itoa(len(tc.wantBody)), rec.Header().Get("Content-Length"))
			assert.Equal(t, string(tc.wantBody), rec.Body.String())
		})
	}
}
