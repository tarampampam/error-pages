package template

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Template struct {
	content string
}

type Templates struct {
	httpClient httpClient
	mu         sync.Mutex
	templates  map[string]Template
}

// Load loads a template from the specified source, which can be either a URL, a file path, or a raw template string.
// The loaded template is stored in the Templates struct for later retrieval and rendering.
func (t *Templates) Load(ctx context.Context, src string) error {
	src = strings.TrimSpace(src)

	if src == "" {
		return fmt.Errorf("source cannot be empty")
	}

	// determine the source type (URL, file path, or raw template string)
	if u, uErr := url.Parse(src); uErr == nil && u.Scheme != "" && u.Host != "" {
		// source is a URL
	} else if filepath.IsAbs(src) || strings.HasPrefix(src, ".") {
	}
}

const templateSizeLimit = 10 << 20 // 10 MB

func (t *Templates) loadFromURL(ctx context.Context, url string) ([]byte, error) {
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if reqErr != nil {
		return nil, reqErr
	}

	resp, respErr := t.httpClient.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.ContentLength > 0 && resp.ContentLength > templateSizeLimit {
		return nil, fmt.Errorf("content length exceeds limit: %d bytes", resp.ContentLength)
	}

	return io.ReadAll(io.LimitReader(resp.Body, templateSizeLimit))
}

func (t *Templates) loadFromFile(path string) ([]byte, error) {
	stat, statErr := os.Stat(path)
	if statErr != nil {
		return nil, statErr
	}

	if stat.Size() > templateSizeLimit {
		return nil, fmt.Errorf("file size exceeds limit: %d bytes", stat.Size())
	}

	f, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	defer func() { _ = f.Close() }()

	return io.ReadAll(io.LimitReader(f, templateSizeLimit))
}
