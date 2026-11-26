package error_page_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/formats"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/error_page"
	"gh.tarampamp.am/error-pages/v4/internal/logger"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

// A real server is used because httptest.ResponseRecorder does not behave like
// net/http for statuses that cannot be sent.
func TestNewSameStatusWithCodesBelow100(t *testing.T) {
	t.Parallel()

	tmpl := mustTemplate(t, `{{.StatusCode}}`)
	server := httptest.NewServer(error_page.New(
		logger.NewNop(),
		404,
		true,
		nil,
		noDesc,
		func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
		false,
		false,
		"",
		nil,
	))
	t.Cleanup(server.Close)

	for path, wantStatus := range map[string]int{
		"/13":  http.StatusOK,
		"/99":  http.StatusOK,
		"/404": http.StatusNotFound,
	} {
		t.Run(path, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+path, nil)
			assert.NoError(t, err)

			resp, err := server.Client().Do(req)
			assert.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, wantStatus, resp.StatusCode)
		})
	}
}
