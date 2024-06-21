package healthcheck_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/appmeta"
	"gh.tarampamp.am/error-pages/internal/cli/healthcheck"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestHealthChecker_CheckSuccess(t *testing.T) {
	t.Parallel()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "foobar:123/healthz", req.URL.String())
		assert.Equal(t, fmt.Sprintf("ErrorPages/%s (HealthCheck)", appmeta.Version()), req.Header.Get("User-Agent"))

		return &http.Response{
			Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
			StatusCode: http.StatusOK,
		}, nil
	}

	assert.NoError(t, healthcheck.NewHTTPHealthChecker(
		healthcheck.WithHttpClient(httpMock),
	).Check(context.Background(), "foobar:123"))
}

func TestHealthChecker_CheckFail(t *testing.T) {
	t.Parallel()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "foobar:123/foo", req.URL.String())

		return &http.Response{
			Body:       http.NoBody,
			StatusCode: http.StatusBadGateway,
		}, nil
	}

	var err = healthcheck.NewHTTPHealthChecker(
		healthcheck.WithHttpClient(httpMock),
		healthcheck.WithLiveEndpoint("foo"),
	).Check(context.Background(), "foobar:123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong status code [502]")
}

func TestHealthChecker_ClientDoError(t *testing.T) {
	t.Parallel()

	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		return nil, assert.AnError
	}

	var err = healthcheck.NewHTTPHealthChecker(
		healthcheck.WithHttpClient(httpMock),
		healthcheck.WithLiveEndpoint("foo"),
	).Check(context.Background(), "foobar:123")

	assert.ErrorIs(t, err, assert.AnError)
}

func TestHTTPHealthChecker_CheckNormalize(t *testing.T) {
	t.Parallel()

	for name, _tc := range map[string]struct {
		giveBaseURL string
		giveLive    string
		wantURL     string
	}{
		"no-live": {
			giveBaseURL: "foobar:123",
			wantURL:     "foobar:123",
		},
		"live with slash": {
			giveBaseURL: "foobar:123",
			giveLive:    "/foo",
			wantURL:     "foobar:123/foo",
		},
		"live without slash": {
			giveBaseURL: "foobar:123",
			giveLive:    "foo",
			wantURL:     "foobar:123/foo",
		},
		"base with slash": {
			giveBaseURL: "foobar:123/",
			giveLive:    "foo",
			wantURL:     "foobar:123/foo",
		},
		"all of slashes": {
			giveBaseURL: "foobar:123/",
			giveLive:    "/foo",
			wantURL:     "foobar:123/foo",
		},
	} {
		tc := _tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, tc.wantURL, req.URL.String())

				return &http.Response{
					Body:       http.NoBody,
					StatusCode: http.StatusOK,
				}, nil
			}

			require.NoError(t, healthcheck.NewHTTPHealthChecker(
				healthcheck.WithHttpClient(httpMock),
				healthcheck.WithLiveEndpoint(tc.giveLive),
			).Check(context.Background(), tc.giveBaseURL))
		})
	}
}
