package tploader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
)

// HTTPClient is the interface for making HTTP requests, used by [FetchContentFromURL].
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type options struct {
	// MaxTemplateSize defines the maximum allowed size for template content in bytes.
	MaxTemplateSize int

	// httpClient allows using a custom HTTP client for fetching remote templates, which can be useful for testing
	// or advanced configurations.
	httpClient HTTPClient
}

// Option allows to configure the behavior of template content loading.
type Option func(*options)

// WithMaxTemplateSize sets the maximum allowed size for template content in bytes.
func WithMaxTemplateSize(size int) Option {
	return func(o *options) { o.MaxTemplateSize = size }
}

// WithHTTPClient allows setting a custom HTTP client for fetching remote templates.
func WithHTTPClient(client HTTPClient) Option {
	if client == nil {
		return func(_ *options) {} // no-op to avoid a nil pointer dereference inside FetchContentFromURL
	}

	return func(o *options) { o.httpClient = client }
}

// newOptions creates an options struct with default values and applies any provided Option functions to it.
func newOptions(opts ...Option) options {
	const (
		defaultMaxTemplateSize       = 5 * 1024 * 1024 // 5 MB
		defaultClientTimeout         = 30 * time.Second
		defaultTLSHandshakeTimeout   = 10 * time.Second
		defaultResponseHeaderTimeout = 10 * time.Second
		defaultIdleConnTimeout       = 90 * time.Second
	)

	o := options{
		MaxTemplateSize: defaultMaxTemplateSize,
		httpClient: &http.Client{
			Timeout: defaultClientTimeout,
			Transport: &http.Transport{
				ForceAttemptHTTP2:     true,
				TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
				ResponseHeaderTimeout: defaultResponseHeaderTimeout,
				IdleConnTimeout:       defaultIdleConnTimeout,
			},
		},
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// LoadTemplateContent attempts to load the template content from the provided source, which can be a URL or
// a file path.
// Since we don't know what exactly source is, we will try:
//
//   - to treat it as a URL first
//   - if that fails, it still can be a file path or a literal template content
//   - if it's a file path, we will read the content of the file
//   - if both checks fail, we will treat the source string itself as the template content
func LoadTemplateContent(ctx context.Context, source string, opts ...Option) (string, error) {
	if IsURL(source) {
		d, err := FetchContentFromURL(ctx, source, opts...)
		if err != nil {
			return "", fmt.Errorf("fetch template from URL: %w", err)
		}

		if len(d) == 0 {
			return "", errors.New("empty content from URL")
		}

		return string(d), nil
	}

	if IsFilePath(source) {
		d, err := ReadContentFromFile(source, opts...)
		if err != nil {
			return "", fmt.Errorf("read template from file: %w", err)
		}

		if len(d) == 0 {
			return "", errors.New("empty content from file")
		}

		return string(d), nil
	}

	// this is not url nor file path, so threat it as a literal template content
	return source, nil
}

// IsURL checks if the provided string is a valid http(s) URL with a host.
func IsURL(s string) bool {
	u, err := url.ParseRequestURI(strings.TrimSpace(s))
	if err != nil {
		return false
	}

	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// FetchContentFromURL retrieves the content from the specified URL with a timeout and returns it as a byte slice.
func FetchContentFromURL(ctx context.Context, urlStr string, opts ...Option) ([]byte, error) {
	o := newOptions(opts...)

	req, rErr := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(urlStr), http.NoBody)
	if rErr != nil {
		return nil, rErr
	}

	req.Header.Set("User-Agent", "error-pages/"+appmeta.Version())

	resp, cErr := o.httpClient.Do(req)
	if cErr != nil {
		return nil, cErr
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch content: non-2xx status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(o.MaxTemplateSize)+1))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if len(bodyBytes) > o.MaxTemplateSize {
		return nil, fmt.Errorf("response exceeds %d bytes", o.MaxTemplateSize)
	}

	return bodyBytes, nil
}

// IsFilePath checks if the provided string is a valid file path that points to a regular file.
func IsFilePath(s string) bool {
	info, err := os.Stat(strings.TrimSpace(s))

	return err == nil && info.Mode().IsRegular()
}

// ReadContentFromFile reads the content of the file at the specified path and returns it as a byte slice.
func ReadContentFromFile(path string, opts ...Option) ([]byte, error) {
	o := newOptions(opts...)

	f, err := os.Open(strings.TrimSpace(path))
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	bodyBytes, err := io.ReadAll(io.LimitReader(f, int64(o.MaxTemplateSize)+1))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if len(bodyBytes) > o.MaxTemplateSize {
		return nil, fmt.Errorf("file content exceeds %d bytes", o.MaxTemplateSize)
	}

	return bodyBytes, nil
}
