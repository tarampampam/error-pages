package healthcheck

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gh.tarampamp.am/error-pages/internal/appmeta"
)

type (
	httpClient interface {
		Do(*http.Request) (*http.Response, error)
	}

	// HealthCheckerOption allows you to change some settings of the checker.
	HealthCheckerOption func(*HTTPHealthChecker)
)

// WithHttpClient allows to set http client.
func WithHttpClient(c httpClient) HealthCheckerOption {
	return func(hc *HTTPHealthChecker) { hc.httpClient = c }
}

// WithLiveEndpoint set the endpoint to check.
func WithLiveEndpoint(endpoint string) HealthCheckerOption {
	if len(endpoint) > 0 && endpoint[0] != '/' {
		endpoint = "/" + endpoint
	}

	return func(hc *HTTPHealthChecker) { hc.liveEndpoint = endpoint }
}

// HTTPHealthChecker is HTTP probe checker.
type HTTPHealthChecker struct {
	httpClient   httpClient
	liveEndpoint string
}

var _ checker = (*HTTPHealthChecker)(nil) // ensure that HTTPHealthChecker implements checker interface

func NewHTTPHealthChecker(opts ...HealthCheckerOption) *HTTPHealthChecker {
	const (
		httpClientTimeout = 3 * time.Second
		liveRoute         = "/healthz"
	)

	var c = HTTPHealthChecker{
		httpClient: &http.Client{
			Timeout:   httpClientTimeout,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
		},
		liveEndpoint: liveRoute,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return &c
}

// Check performs HTTP get request.
func (c *HTTPHealthChecker) Check(ctx context.Context, baseURL string) error {
	var endpoint = strings.TrimRight(strings.TrimSpace(baseURL), "/") + c.liveEndpoint

	var req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("ErrorPages/%s (HealthCheck)", appmeta.Version()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	_ = resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("wrong status code [%d] from the live endpoint (%s)", code, endpoint)
	}

	return nil
}
