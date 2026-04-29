package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/httpserver/middleware"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("common cases", func(t *testing.T) {
		t.Parallel()

		var calls []string

		mw := func(name string) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls = append(calls, "before-"+name)

					next.ServeHTTP(w, r)

					calls = append(calls, "after-"+name)
				})
			}
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "handler")
		})

		h := middleware.Apply(handler, mw("A"), mw("B"), mw("C"))

		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		assert.DeepEqual(t, []string{
			"before-A",
			"before-B",
			"before-C",
			"handler",
			"after-C",
			"after-B",
			"after-A",
		}, calls)
	})

	t.Run("skip nil middleware", func(t *testing.T) {
		t.Parallel()

		var called bool

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})

		h := middleware.Apply(handler, nil, nil)

		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		assert.True(t, called)
	})

	t.Run("no middleware", func(t *testing.T) {
		t.Parallel()

		var called bool

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})

		h := middleware.Apply(handler)

		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		assert.True(t, called)
	})

	t.Run("modify request", func(t *testing.T) {
		t.Parallel()

		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-Test", "123")
				next.ServeHTTP(w, r)
			})
		}

		var got string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = r.Header.Get("X-Test")
		})

		h := middleware.Apply(handler, mw)

		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, "123", got)
	})
}
