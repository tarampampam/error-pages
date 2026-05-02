package error_page_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/codes"
	"gh.tarampamp.am/error-pages/v4/internal/formats"
	"gh.tarampamp.am/error-pages/v4/internal/httpserver/handlers/error_page"
	"gh.tarampamp.am/error-pages/v4/internal/logger"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

// mustTemplate parses src as a Go template or fails the test immediately.
func mustTemplate(t *testing.T, src string) *tpl.Template {
	t.Helper()

	tmpl, err := tpl.New(src)
	assert.NoError(t, err)

	return tmpl
}

// noDesc is a code describer that always returns not-found.
func noDesc(uint16) (codes.Description, bool) { return codes.Description{}, false }

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("method gating", func(t *testing.T) {
		t.Parallel()

		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(_ formats.Format) (*tpl.Template, error) { return mustTemplate(t, ""), nil },
			false,
			false,
			"",
		)

		for name, tc := range map[string]struct {
			giveMethod string
			wantStatus int
			wantAllow  string
		}{
			"GET allowed":     {giveMethod: http.MethodGet, wantStatus: http.StatusOK},
			"HEAD allowed":    {giveMethod: http.MethodHead, wantStatus: http.StatusOK},
			"POST rejected":   {giveMethod: http.MethodPost, wantStatus: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
			"PUT rejected":    {giveMethod: http.MethodPut, wantStatus: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
			"DELETE rejected": {giveMethod: http.MethodDelete, wantStatus: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
			"PATCH rejected":  {giveMethod: http.MethodPatch, wantStatus: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(tc.giveMethod, "/404", nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, tc.wantStatus, rec.Code)

				if tc.wantAllow != "" {
					assert.Equal(t, tc.wantAllow, rec.Header().Get("Allow"))
				}
			})
		}
	})

	t.Run("code extraction", func(t *testing.T) {
		t.Parallel()

		// template outputs the status code so we can verify what code was resolved
		tmpl := mustTemplate(t, `{{.StatusCode}}`)

		makeHandler := func(defaultCode uint16) http.Handler {
			return error_page.New(
				logger.NewNop(),
				defaultCode,
				false,
				nil,
				noDesc,
				func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
				false,
				false,
				"",
			)
		}

		for name, tc := range map[string]struct {
			giveDefaultCode uint16
			givePath        string
			giveXCode       string
			wantBody        string
		}{
			"numeric path /404":                  {giveDefaultCode: 500, givePath: "/404", wantBody: "404"},
			"path with html extension /503.html": {giveDefaultCode: 500, givePath: "/503.html", wantBody: "503"},
			"path with json extension /503.json": {giveDefaultCode: 500, givePath: "/503.json", wantBody: "503"},
			"path with sub-segment /404/details": {giveDefaultCode: 500, givePath: "/404/details", wantBody: "404"},
			"root path uses default":             {giveDefaultCode: 418, givePath: "/", wantBody: "418"},
			"text segment uses default":          {giveDefaultCode: 418, givePath: "/notfound", wantBody: "418"},
			"code 1 is valid lower bound":        {giveDefaultCode: 500, givePath: "/1", wantBody: "1"},
			"code 999 is valid upper bound":      {giveDefaultCode: 500, givePath: "/999", wantBody: "999"},
			"code 0 is invalid, uses default":    {giveDefaultCode: 418, givePath: "/0", wantBody: "418"},
			"code 1000 is invalid, uses default": {giveDefaultCode: 418, givePath: "/1000", wantBody: "418"},
			"X-Code fallback for non-numeric path": {
				giveDefaultCode: 500,
				givePath:        "/",
				giveXCode:       "502",
				wantBody:        "502",
			},
			"URL path beats X-Code header": {
				giveDefaultCode: 500,
				givePath:        "/404",
				giveXCode:       "503",
				wantBody:        "404",
			},
			"X-Code longer than 3 chars ignored": {
				giveDefaultCode: 418,
				givePath:        "/",
				giveXCode:       "1000",
				wantBody:        "418",
			},
			"X-Code 0 ignored": {giveDefaultCode: 418, givePath: "/", giveXCode: "0", wantBody: "418"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				if tc.giveXCode != "" {
					req.Header.Set("X-Code", tc.giveXCode)
				}

				rec := httptest.NewRecorder()
				makeHandler(tc.giveDefaultCode).ServeHTTP(rec, req)

				assert.Equal(t, tc.wantBody, rec.Body.String())
			})
		}
	})

	t.Run("format detection", func(t *testing.T) {
		t.Parallel()

		// pre-create per-format templates so we can verify both Content-Type header and body
		jsonTmpl := mustTemplate(t, `json-body`)
		xmlTmpl := mustTemplate(t, `xml-body`)
		htmlTmpl := mustTemplate(t, `html-body`)
		textTmpl := mustTemplate(t, `text-body`)

		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(f formats.Format) (*tpl.Template, error) {
				switch f {
				case formats.JSONFormat:
					return jsonTmpl, nil
				case formats.XMLFormat:
					return xmlTmpl, nil
				case formats.HTMLFormat:
					return htmlTmpl, nil
				case formats.PlainTextFormat:
					return textTmpl, nil
				}

				return textTmpl, nil
			},
			false,
			false,
			"",
		)

		for name, tc := range map[string]struct {
			givePath        string
			giveContentType string
			giveXFormat     string
			giveAccept      string
			wantContentType string
			wantBody        string
		}{
			// URL extension has highest priority
			"extension .json": {
				givePath: "/404.json", wantContentType: "application/json; charset=utf-8", wantBody: "json-body",
			},
			"extension .xml": {
				givePath: "/404.xml", wantContentType: "application/xml; charset=utf-8", wantBody: "xml-body",
			},
			"extension .html": {
				givePath: "/404.html", wantContentType: "text/html; charset=utf-8", wantBody: "html-body",
			},
			"extension .htm": {
				givePath: "/404.htm", wantContentType: "text/html; charset=utf-8", wantBody: "html-body",
			},
			"extension .txt": {
				givePath: "/404.txt", wantContentType: "text/plain; charset=utf-8", wantBody: "text-body",
			},
			"extension beats Accept header": {
				givePath:        "/404.json",
				giveAccept:      "text/html",
				wantContentType: "application/json; charset=utf-8",
				wantBody:        "json-body",
			},

			// Content-Type header (second priority)
			"Content-Type application/json without params": {
				givePath:        "/404",
				giveContentType: "application/json",
				wantContentType: "application/json; charset=utf-8",
				wantBody:        "json-body",
			},
			"Content-Type text/html with charset param": {
				givePath:        "/404",
				giveContentType: "text/html; charset=utf-8",
				wantContentType: "text/html; charset=utf-8",
				wantBody:        "html-body",
			},
			"Content-Type application/xml without params": {
				givePath:        "/404",
				giveContentType: "application/xml",
				wantContentType: "application/xml; charset=utf-8",
				wantBody:        "xml-body",
			},
			"Content-Type text/plain without params": {
				givePath:        "/404",
				giveContentType: "text/plain",
				wantContentType: "text/plain; charset=utf-8",
				wantBody:        "text-body",
			},
			"unknown Content-Type falls through to X-Format": {
				givePath:        "/404",
				giveContentType: "text/css",
				giveXFormat:     "text/html",
				wantContentType: "text/html; charset=utf-8",
				wantBody:        "html-body",
			},
			"unknown Content-Type falls through to Accept": {
				givePath:        "/404",
				giveContentType: "text/css",
				giveAccept:      "application/json",
				wantContentType: "application/json; charset=utf-8",
				wantBody:        "json-body",
			},

			// X-Format header (third priority)
			"X-Format text/html": {
				givePath:        "/404",
				giveXFormat:     "text/html",
				wantContentType: "text/html; charset=utf-8",
				wantBody:        "html-body",
			},
			"X-Format application/json": {
				givePath:        "/404",
				giveXFormat:     "application/json",
				wantContentType: "application/json; charset=utf-8",
				wantBody:        "json-body",
			},
			"X-Format beats Accept": {
				givePath:        "/404",
				giveXFormat:     "text/html",
				giveAccept:      "application/json",
				wantContentType: "text/html; charset=utf-8",
				wantBody:        "html-body",
			},

			// Accept header (fourth priority)
			"Accept text/html": {
				givePath:        "/404",
				giveAccept:      "text/html",
				wantContentType: "text/html; charset=utf-8",
				wantBody:        "html-body",
			},
			"Accept application/json": {
				givePath:        "/404",
				giveAccept:      "application/json",
				wantContentType: "application/json; charset=utf-8",
				wantBody:        "json-body",
			},
			"Accept application/xhtml+xml maps to XML": {
				givePath:        "/404",
				giveAccept:      "application/xhtml+xml",
				wantContentType: "application/xml; charset=utf-8",
				wantBody:        "xml-body",
			},
			"Accept q-weight: JSON higher weight wins": {
				givePath:        "/404",
				giveAccept:      "text/html;q=0.5,application/json;q=0.9",
				wantContentType: "application/json; charset=utf-8",
				wantBody:        "json-body",
			},
			"Accept q=0 entry excluded, remaining wins": {
				givePath:        "/404",
				giveAccept:      "text/html;q=0,application/json;q=0.5",
				wantContentType: "application/json; charset=utf-8",
				wantBody:        "json-body",
			},
			"Accept all q=0 falls back to default plain": {
				givePath:        "/404",
				giveAccept:      "text/html;q=0,application/json;q=0",
				wantContentType: "text/plain; charset=utf-8",
				wantBody:        "text-body",
			},
			"Accept wildcard */* falls back to default plain": {
				givePath:        "/404",
				giveAccept:      "*/*",
				wantContentType: "text/plain; charset=utf-8",
				wantBody:        "text-body",
			},

			// No hints at all - default to plain text
			"no format hints - default plain text": {
				givePath:        "/404",
				wantContentType: "text/plain; charset=utf-8",
				wantBody:        "text-body",
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				if tc.giveContentType != "" {
					req.Header.Set("Content-Type", tc.giveContentType)
				}

				if tc.giveXFormat != "" {
					req.Header.Set("X-Format", tc.giveXFormat)
				}

				if tc.giveAccept != "" {
					req.Header.Set("Accept", tc.giveAccept)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, tc.wantContentType, rec.Header().Get("Content-Type"))
				assert.Equal(t, tc.wantBody, rec.Body.String())
			})
		}
	})

	t.Run("respond same status", func(t *testing.T) {
		t.Parallel()

		tmpl := mustTemplate(t, `ok`)

		for name, tc := range map[string]struct {
			giveRespondSameStatus bool
			giveDefaultCode       uint16
			givePath              string
			wantStatus            int
		}{
			"false: always 200 for 503":     {false, 404, "/503", http.StatusOK},
			"false: always 200 for 404":     {false, 404, "/404", http.StatusOK},
			"true: reflects 503":            {true, 404, "/503", http.StatusServiceUnavailable},
			"true: reflects 404":            {true, 404, "/404", http.StatusNotFound},
			"true: uses default code for /": {true, 418, "/", http.StatusTeapot},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				h := error_page.New(
					logger.NewNop(),
					tc.giveDefaultCode,
					tc.giveRespondSameStatus,
					nil,
					noDesc,
					func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
					false,
					false,
					"",
				)

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, tc.wantStatus, rec.Code)
			})
		}
	})

	t.Run("response headers", func(t *testing.T) {
		t.Parallel()

		tmpl := mustTemplate(t, `ok`)

		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
			false,
			false,
			"",
		)

		for name, tc := range map[string]struct {
			givePath       string
			wantRetryAfter string
		}{
			"404 - no retry-after":                        {givePath: "/404"},
			"403 - no retry-after":                        {givePath: "/403"},
			"408 Request Timeout - retry-after 120":       {givePath: "/408", wantRetryAfter: "120"},
			"425 Too Early - retry-after 120":             {givePath: "/425", wantRetryAfter: "120"},
			"429 Too Many Requests - retry-after 120":     {givePath: "/429", wantRetryAfter: "120"},
			"500 Internal Server Error - retry-after 120": {givePath: "/500", wantRetryAfter: "120"},
			"502 Bad Gateway - retry-after 120":           {givePath: "/502", wantRetryAfter: "120"},
			"503 Service Unavailable - retry-after 120":   {givePath: "/503", wantRetryAfter: "120"},
			"504 Gateway Timeout - retry-after 120":       {givePath: "/504", wantRetryAfter: "120"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, "noindex, nofollow, nosnippet, noarchive", rec.Header().Get("X-Robots-Tag"))
				assert.Equal(t, tc.wantRetryAfter, rec.Header().Get("Retry-After"))
			})
		}
	})

	t.Run("proxy headers", func(t *testing.T) {
		t.Parallel()

		tmpl := mustTemplate(t, `ok`)

		for name, tc := range map[string]struct {
			giveProxyHeaders   []string
			giveRequestHeaders map[string]string
			wantRespHeaders    map[string]string
		}{
			"configured header is forwarded": {
				giveProxyHeaders:   []string{"X-Custom"},
				giveRequestHeaders: map[string]string{"X-Custom": "my-value"},
				wantRespHeaders:    map[string]string{"X-Custom": "my-value"},
			},
			"non-configured header is not forwarded": {
				giveProxyHeaders:   []string{"X-Custom"},
				giveRequestHeaders: map[string]string{"X-Other": "other-value"},
				wantRespHeaders:    map[string]string{"X-Other": ""},
			},
			"empty request header value is not forwarded": {
				giveProxyHeaders:   []string{"X-Custom"},
				giveRequestHeaders: map[string]string{},
				wantRespHeaders:    map[string]string{"X-Custom": ""},
			},
			"multiple configured headers forwarded": {
				giveProxyHeaders:   []string{"X-A", "X-B"},
				giveRequestHeaders: map[string]string{"X-A": "val-a", "X-B": "val-b"},
				wantRespHeaders:    map[string]string{"X-A": "val-a", "X-B": "val-b"},
			},
			"nil proxy list forwards nothing": {
				giveProxyHeaders:   nil,
				giveRequestHeaders: map[string]string{"X-Custom": "value"},
				wantRespHeaders:    map[string]string{"X-Custom": ""},
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				h := error_page.New(
					logger.NewNop(),
					404,
					false,
					tc.giveProxyHeaders,
					noDesc,
					func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
					false,
					false,
					"",
				)

				req := httptest.NewRequest(http.MethodGet, "/404", nil)
				for k, v := range tc.giveRequestHeaders {
					req.Header.Set(k, v)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				for header, want := range tc.wantRespHeaders {
					assert.Equal(t, want, rec.Header().Get(header))
				}
			})
		}
	})

	t.Run("code description", func(t *testing.T) {
		t.Parallel()

		// template outputs both fields so we can verify which description was resolved
		tmpl := mustTemplate(t, `{{.Message}}|{{.Description}}`)

		for name, tc := range map[string]struct {
			giveDescriber func(uint16) (codes.Description, bool)
			givePath      string
			wantContains  []string
		}{
			"describer result is used": {
				giveDescriber: func(_ uint16) (codes.Description, bool) {
					return codes.Description{Short: "Custom Short", Full: "Custom Full"}, true
				},
				givePath:     "/404",
				wantContains: []string{"Custom Short", "Custom Full"},
			},
			"stdlib fallback for known code 404": {
				giveDescriber: noDesc,
				givePath:      "/404",
				wantContains:  []string{"Not Found"},
			},
			"stdlib fallback for known code 503": {
				giveDescriber: noDesc,
				givePath:      "/503",
				wantContains:  []string{"Service Unavailable"},
			},
			"unknown code uses hardcoded fallback string": {
				giveDescriber: noDesc,
				givePath:      "/600", // no stdlib text exists for 600
				wantContains:  []string{"Unknown Status Code"},
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				h := error_page.New(
					logger.NewNop(),
					404,
					false,
					nil,
					tc.giveDescriber,
					func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
					false,
					false,
					"",
				)

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Contains(t, rec.Body.String(), tc.wantContains...)
			})
		}
	})

	t.Run("nil template", func(t *testing.T) {
		t.Parallel()

		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(_ formats.Format) (*tpl.Template, error) {
				return nil, nil //nolint:nilnil // no template is the scenario under test
			},
			false,
			false,
			"",
		)

		for name, tc := range map[string]struct {
			givePath     string
			wantContains string
		}{
			"plain text error message": {givePath: "/404.txt", wantContains: "No template available"},
			"JSON error has error key": {givePath: "/404.json", wantContains: `"error"`},
			"XML error has error tag":  {givePath: "/404.xml", wantContains: "<error>"},
			"HTML error has doctype":   {givePath: "/404.html", wantContains: "<!DOCTYPE html>"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Contains(t, rec.Body.String(), tc.wantContains)
			})
		}
	})

	t.Run("templater returns error", func(t *testing.T) {
		t.Parallel()

		templaterErr := errors.New("template unavailable")

		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(_ formats.Format) (*tpl.Template, error) { return nil, templaterErr },
			false,
			false,
			"",
		)

		for name, tc := range map[string]struct {
			givePath     string
			wantContains string
		}{
			"plain text error on templater failure": {givePath: "/404.txt", wantContains: "Failed to get the template"},
			"JSON error on templater failure":       {givePath: "/404.json", wantContains: `"error"`},
			"XML error on templater failure":        {givePath: "/404.xml", wantContains: "<error>"},
			"HTML error on templater failure":       {givePath: "/404.html", wantContains: "<!DOCTYPE html>"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Contains(t, rec.Body.String(), tc.wantContains)
			})
		}
	})

	t.Run("template render error", func(t *testing.T) {
		t.Parallel()

		// {{.NonExistentField}} parses fine but fails at render time - field does not exist in tpl.Data
		badTmpl := mustTemplate(t, `{{.NonExistentField}}`)

		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(_ formats.Format) (*tpl.Template, error) { return badTmpl, nil },
			false,
			false,
			"",
		)

		for name, tc := range map[string]struct {
			givePath     string
			wantContains string
		}{
			"plain text error on render failure": {givePath: "/404.txt", wantContains: "Failed to render"},
			"JSON error on render failure":       {givePath: "/404.json", wantContains: `"error"`},
			"XML error on render failure":        {givePath: "/404.xml", wantContains: "<error>"},
			"HTML error on render failure":       {givePath: "/404.html", wantContains: "<!DOCTYPE html>"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(http.MethodGet, tc.givePath, nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Contains(t, rec.Body.String(), tc.wantContains)
			})
		}
	})

	t.Run("show details", func(t *testing.T) {
		t.Parallel()

		// template renders all detail fields in a predictable order
		const tplSrc = `{{.OriginalURI}},{{.Namespace}},{{.IngressName}},` +
			`{{.ServiceName}},{{.ServicePort}},{{.RequestID}},{{.ForwardedFor}},{{.Host}}`

		tmpl := mustTemplate(t, tplSrc)

		t.Run("true populates ingress-nginx fields", func(t *testing.T) {
			t.Parallel()

			h := error_page.New(
				logger.NewNop(),
				404,
				false,
				nil,
				noDesc,
				func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
				true,
				false,
				"",
			)

			req := httptest.NewRequest(http.MethodGet, "/500", nil)
			req.Header.Set("X-Original-Uri", "/app/path")
			req.Header.Set("X-Namespace", "production")
			req.Header.Set("X-Ingress-Name", "my-ingress")
			req.Header.Set("X-Service-Name", "my-service")
			req.Header.Set("X-Service-Port", "8080")
			req.Header.Set("X-Request-Id", "abc-123")
			req.Header.Set("X-Forwarded-For", "1.2.3.4")
			req.Header.Set("Host", "example.com")

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			assert.Contains(t, rec.Body.String(),
				"/app/path", "production", "my-ingress",
				"my-service", "8080", "abc-123",
				"1.2.3.4", "example.com",
			)
		})

		t.Run("false leaves all detail fields empty", func(t *testing.T) {
			t.Parallel()

			h := error_page.New(
				logger.NewNop(),
				404,
				false,
				nil,
				noDesc,
				func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
				false,
				false,
				"",
			)

			req := httptest.NewRequest(http.MethodGet, "/500", nil)
			req.Header.Set("X-Original-Uri", "/app/path")
			req.Header.Set("X-Namespace", "production")

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			// 8 empty fields separated by 7 commas
			assert.Equal(t, ",,,,,,,", rec.Body.String())
		})
	})

	t.Run("GET writes body HEAD omits it", func(t *testing.T) {
		t.Parallel()

		const tplBody = "hello-from-template"

		tmpl := mustTemplate(t, tplBody)
		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
			false,
			false,
			"",
		)

		wantLen := strconv.Itoa(len(tplBody))

		for name, tc := range map[string]struct {
			giveMethod string
			wantBody   string
		}{
			"GET writes body and sets Content-Length": {giveMethod: http.MethodGet, wantBody: tplBody},
			"HEAD omits body but sets Content-Length": {giveMethod: http.MethodHead, wantBody: ""},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(tc.giveMethod, "/404", nil)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, wantLen, rec.Header().Get("Content-Length"))
				assert.Equal(t, tc.wantBody, rec.Body.String())
			})
		}
	})

	t.Run("gzip compression", func(t *testing.T) {
		t.Parallel()

		const tplBody = "hello-gzip"

		tmpl := mustTemplate(t, tplBody)

		// pre-compute the expected gzip output length to verify Content-Length without re-implementing compression
		var wantGzipLen int
		{
			var buf bytes.Buffer

			gw := gzip.NewWriter(&buf)
			_, _ = gw.Write([]byte(tplBody))
			_ = gw.Close()
			wantGzipLen = buf.Len()
		}

		h := error_page.New(
			logger.NewNop(),
			404,
			false,
			nil,
			noDesc,
			func(_ formats.Format) (*tpl.Template, error) { return tmpl, nil },
			false,
			false,
			"",
		)

		for name, tc := range map[string]struct {
			giveMethod         string
			giveAcceptEncoding string
			wantEncoding       string
			wantVary           string
			wantCompressed     bool
		}{
			"GET accepts gzip: body compressed": {
				giveMethod: http.MethodGet, giveAcceptEncoding: "gzip",
				wantEncoding: "gzip", wantVary: "Accept-Encoding", wantCompressed: true,
			},
			"GET accepts gzip among others: body compressed": {
				giveMethod: http.MethodGet, giveAcceptEncoding: "deflate, gzip",
				wantEncoding: "gzip", wantVary: "Accept-Encoding", wantCompressed: true,
			},
			"GET no Accept-Encoding: plain body": {
				giveMethod: http.MethodGet, giveAcceptEncoding: "",
			},
			"GET non-gzip encoding: plain body": {
				giveMethod: http.MethodGet, giveAcceptEncoding: "deflate",
			},
			"HEAD accepts gzip: headers set, body empty": {
				giveMethod: http.MethodHead, giveAcceptEncoding: "gzip",
				wantEncoding: "gzip", wantVary: "Accept-Encoding", wantCompressed: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(tc.giveMethod, "/404", nil)
				if tc.giveAcceptEncoding != "" {
					req.Header.Set("Accept-Encoding", tc.giveAcceptEncoding)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				assert.Equal(t, tc.wantEncoding, rec.Header().Get("Content-Encoding"))
				assert.Equal(t, tc.wantVary, rec.Header().Get("Vary"))

				if tc.wantCompressed {
					assert.Equal(t, strconv.Itoa(wantGzipLen), rec.Header().Get("Content-Length"))
				} else {
					assert.Equal(t, strconv.Itoa(len(tplBody)), rec.Header().Get("Content-Length"))
				}

				if tc.giveMethod == http.MethodHead {
					assert.Equal(t, "", rec.Body.String())

					return
				}

				if tc.wantCompressed {
					gr, grErr := gzip.NewReader(rec.Body)
					assert.NoError(t, grErr)

					var decompressed bytes.Buffer

					_, cpErr := io.Copy(&decompressed, io.LimitReader(gr, int64(len(tplBody)+1))) //nolint:gosec // bounded by known test content size
					assert.NoError(t, cpErr)
					assert.NoError(t, gr.Close())
					assert.Equal(t, tplBody, decompressed.String())
				} else {
					assert.Equal(t, tplBody, rec.Body.String())
				}
			})
		}

		t.Run("empty body not compressed", func(t *testing.T) {
			t.Parallel()

			emptyTmpl := mustTemplate(t, "")
			hEmpty := error_page.New(
				logger.NewNop(),
				404,
				false,
				nil,
				noDesc,
				func(_ formats.Format) (*tpl.Template, error) { return emptyTmpl, nil },
				false,
				false,
				"",
			)

			req := httptest.NewRequest(http.MethodGet, "/404", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			rec := httptest.NewRecorder()
			hEmpty.ServeHTTP(rec, req)

			assert.Equal(t, "", rec.Header().Get("Content-Encoding"))
			assert.Equal(t, "", rec.Header().Get("Vary"))
			assert.Equal(t, "0", rec.Header().Get("Content-Length"))
			assert.Equal(t, "", rec.Body.String())
		})
	})
}
