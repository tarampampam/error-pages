package error_page_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/http/handlers/error_page"
	"gh.tarampamp.am/error-pages/internal/http/httptest"
	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		giveConfig  func() *config.Config
		giveUrl     string
		giveHeaders map[string]string

		wantStatusCode   int
		wantHeaders      map[string]string
		wantBodyIncludes []string
	}{
		"common, plain text": {
			giveConfig:  func() *config.Config { cfg := config.New(); return &cfg },
			giveUrl:     "http://testing/",
			giveHeaders: map[string]string{"Content-Type": "text/plain"},

			wantStatusCode:   http.StatusOK,
			wantHeaders:      map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			wantBodyIncludes: []string{"Error 404", "Not Found"},
		},
		"common, html": {
			giveConfig: func() *config.Config {
				cfg := config.New()

				cfg.TemplateName = "ghost"

				return &cfg
			},
			giveUrl:     "http://testing/",
			giveHeaders: map[string]string{"X-Format": "text/html", "X-Code": "407"},

			wantStatusCode: http.StatusOK,
			wantHeaders:    map[string]string{"Content-Type": "text/html; charset=utf-8"},
			wantBodyIncludes: []string{
				"<!doctype html>",
				"<title>407: Proxy Authentication Required",
				"Proxy Authentication Required",
			},
		},
		"common, json": {
			giveConfig: func() *config.Config {
				cfg := config.New()

				cfg.RespondWithSameHTTPCode = true

				return &cfg
			},
			giveUrl:     "http://testing/503.html?rnd=123",
			giveHeaders: map[string]string{"Accept": "application/json", "X-FooBar": "baz"},

			wantStatusCode: http.StatusServiceUnavailable,
			wantHeaders: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
				"X-FooBar":     "", // is not in the list of proxy headers
			},
			wantBodyIncludes: []string{"503", "Service Unavailable"},
		},
		"common, xml": {
			giveConfig: func() *config.Config {
				cfg := config.New()

				cfg.ProxyHeaders = append(cfg.ProxyHeaders, "X-FooBar")

				return &cfg
			},
			giveUrl:     "http://testing/500",
			giveHeaders: map[string]string{"Accept": "application/xml", "X-FooBar": "baz"},

			wantStatusCode: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type": "application/xml; charset=utf-8",
				"X-FooBar":     "baz",
			},
			wantBodyIncludes: []string{"500", "Internal Server Error"},
		},
		"show details": {
			giveConfig: func() *config.Config {
				cfg := config.New()

				cfg.ShowDetails = true

				return &cfg
			},
			giveUrl: "http://example.com/503",
			giveHeaders: map[string]string{
				"Accept":          "application/json",
				"X-Original-URI":  "/foo/bar",
				"X-Namespace":     "some-Namespace",
				"X-Ingress-Name":  "ingress-name",
				"X-Service-Name":  "service-name",
				"X-Service-Port":  "666",
				"X-Request-ID":    "req-id-777",
				"X-Forwarded-For": "123.123.123.123:12312",
			},

			wantStatusCode: http.StatusOK,
			wantHeaders:    map[string]string{"Content-Type": "application/json; charset=utf-8"},
			wantBodyIncludes: []string{
				"503",
				"Service Unavailable",
				"details",
				"/foo/bar",
				"some-Namespace",
				"ingress-name",
				"service-name",
				"666",
				"req-id-777",
				"123.123.123.123:12312",
				"example.com",
			},
		},
		"fallback to StatusText if code is not found": {
			giveConfig: func() *config.Config {
				cfg := config.New()

				cfg.Codes = config.Codes{}

				return &cfg
			},
			giveUrl:     "http://testing/100",
			giveHeaders: map[string]string{"Accept": "application/json"},

			wantStatusCode:   http.StatusOK,
			wantHeaders:      map[string]string{"Content-Type": "application/json; charset=utf-8"},
			wantBodyIncludes: []string{"100", "Continue"},
		},
		"unknown code": {
			giveConfig: func() *config.Config {
				cfg := config.New()

				cfg.Codes = config.Codes{}

				return &cfg
			},
			giveUrl:     "http://testing/1",
			giveHeaders: map[string]string{"Accept": "application/json"},

			wantStatusCode:   http.StatusOK,
			wantHeaders:      map[string]string{"Content-Type": "application/json; charset=utf-8"},
			wantBodyIncludes: []string{"1", "Unknown Status Code"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var handler, closeCache = error_page.New(tt.giveConfig(), logger.NewNop())
			defer closeCache()

			req, reqErr := http.NewRequest(http.MethodGet, tt.giveUrl, http.NoBody)
			require.NoError(t, reqErr)

			for k, v := range tt.giveHeaders {
				req.Header.Set(k, v)
			}

			httptest.HandleFastRequest(t, handler, req, func(status int, body string, headers http.Header) {
				assert.Equal(t, tt.wantStatusCode, status)

				for hName, hWant := range tt.wantHeaders {
					for hGot := range headers {
						if hGot == hName {
							assert.Contains(t, hWant, headers.Get(hGot))
						}
					}
				}

				for _, wantBodyInclude := range tt.wantBodyIncludes {
					assert.Contains(t, body, wantBodyInclude)
				}
			})
		})
	}
}

func TestRotationModeOnEachRequest(t *testing.T) {
	t.Parallel()

	var cfg = config.New()

	cfg.RotationMode = config.RotationModeRandomOnEachRequest
	cfg.Templates = map[string]string{
		"foo": "foo",
		"bar": "bar",
	}

	var (
		lastResponseBody string
		changedTimes     int

		handler, closeCache = error_page.New(&cfg, logger.NewNop())
	)

	defer func() { closeCache(); closeCache(); closeCache() }() // multiple calls should not panic

	for range 300 {
		req, reqErr := http.NewRequest(http.MethodGet, "http://testing/", http.NoBody)
		require.NoError(t, reqErr)

		req.Header.Set("Accept", "text/html")

		httptest.HandleFastRequest(t, handler, req, func(status int, body string, headers http.Header) {
			if lastResponseBody != body {
				changedTimes++
				lastResponseBody = body
			}
		})
	}

	assert.True(t, changedTimes > 30, "the template should be changed at least 30 times")
}
