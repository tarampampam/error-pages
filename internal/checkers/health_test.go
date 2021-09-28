package checkers_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/checkers"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestHealthChecker_CheckSuccess(t *testing.T) {
	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "http://127.0.0.1:123/health/live", req.URL.String())
		assert.Equal(t, "HealthChecker/internal", req.Header.Get("User-Agent"))

		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			StatusCode: http.StatusOK,
		}, nil
	}

	checker := checkers.NewHealthChecker(context.Background(), httpMock)

	assert.NoError(t, checker.Check(123))
}

func TestHealthChecker_CheckFail(t *testing.T) {
	var httpMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			StatusCode: http.StatusBadGateway,
		}, nil
	}

	checker := checkers.NewHealthChecker(context.Background(), httpMock)

	err := checker.Check(123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong status code")
}
