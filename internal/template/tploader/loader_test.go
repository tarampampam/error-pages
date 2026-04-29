package tploader_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/template/tploader"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

type mockHTTPClient func(*http.Request) (*http.Response, error)

func (m mockHTTPClient) Do(req *http.Request) (*http.Response, error) { return m(req) }

// mockRoundTripper implements http.RoundTripper via a plain function.
// Use it with &http.Client{Transport: ...} when redirect-following behavior is needed.
type mockRoundTripper func(*http.Request) (*http.Response, error)

func (m mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return m(req) }

func fakeResp(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestIsURL(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		give string
		want bool
	}{
		"http":                {give: "http://example.com", want: true},
		"https with path":     {give: "https://example.com/template.html", want: true},
		"https with port":     {give: "https://example.com:8080/path?q=1", want: true},
		"ftp scheme":          {give: "ftp://example.com/file.txt", want: false},
		"whitespace padded":   {give: "  https://example.com  ", want: true},
		"absolute file path":  {give: "/etc/hosts", want: false},
		"relative path":       {give: "templates/error.html", want: false},
		"empty":               {give: "", want: false},
		"scheme without host": {give: "http://", want: false},
		"no scheme":           {give: "example.com/path", want: false},
		"whitespace only":     {give: "   ", want: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tploader.IsURL(tc.give))
		})
	}
}

func TestIsFilePath(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		give string
		want bool
	}{
		"nonexistent path": {give: "/nonexistent/path/file.txt", want: false},
		"empty string":     {give: "", want: false},
		"http URL":         {give: "http://example.com", want: false},
		"whitespace only":  {give: "   ", want: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tploader.IsFilePath(tc.give))
		})
	}

	t.Run("regular file", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.txt")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		assert.True(t, tploader.IsFilePath(f.Name()))
	})

	t.Run("directory", func(t *testing.T) {
		t.Parallel()

		assert.False(t, tploader.IsFilePath(t.TempDir()))
	})

	t.Run("whitespace padded path", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.txt")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		assert.True(t, tploader.IsFilePath("  "+f.Name()+"  "))
	})
}

func TestFetchContentFromURL(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveClient      tploader.HTTPClient
		giveOpts        []tploader.Option
		wantBody        string
		wantErrContains string
	}{
		"200 with body": {
			giveClient: mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
				return fakeResp(http.StatusOK, "template content"), nil
			}),
			wantBody: "template content",
		},
		"201 with body": {
			giveClient: mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
				return fakeResp(http.StatusCreated, "created"), nil
			}),
			wantBody: "created",
		},
		"200 empty body": {
			giveClient: mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
				return fakeResp(http.StatusOK, ""), nil
			}),
			wantBody: "",
		},
		"redirect followed": {
			// Uses *http.Client with a mock transport so that redirect-following is handled by the stdlib.
			giveClient: &http.Client{
				Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
					if req.URL.Path == "/" {
						return &http.Response{
							StatusCode: http.StatusMovedPermanently,
							Header:     http.Header{"Location": {"http://fake.test/final"}},
							Body:       io.NopCloser(strings.NewReader("")),
						}, nil
					}

					return fakeResp(http.StatusOK, "final content"), nil
				}),
			},
			wantBody: "final content",
		},
		"404": {
			giveClient: mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
				return fakeResp(http.StatusNotFound, ""), nil
			}),
			wantErrContains: "non-2xx status code: 404",
		},
		"500": {
			giveClient: mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
				return fakeResp(http.StatusInternalServerError, ""), nil
			}),
			wantErrContains: "non-2xx status code: 500",
		},
		"body exactly at size limit": {
			giveClient: mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
				return fakeResp(http.StatusOK, strings.Repeat("x", 100)), nil
			}),
			giveOpts: []tploader.Option{tploader.WithMaxTemplateSize(100)},
			wantBody: strings.Repeat("x", 100),
		},
		"body one byte over size limit": {
			giveClient: mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
				return fakeResp(http.StatusOK, strings.Repeat("x", 101)), nil
			}),
			giveOpts:        []tploader.Option{tploader.WithMaxTemplateSize(100)},
			wantErrContains: "exceeds 100 bytes",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := append([]tploader.Option{tploader.WithHTTPClient(tc.giveClient)}, tc.giveOpts...)
			got, err := tploader.FetchContentFromURL(t.Context(), "http://fake.test/", opts...)

			if tc.wantErrContains != "" {
				assert.ErrorContains(t, err, tc.wantErrContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBody, string(got))
			}
		})
	}

	t.Run("canceled context", func(t *testing.T) {
		t.Parallel()

		client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
			return nil, req.Context().Err()
		})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := tploader.FetchContentFromURL(ctx, "http://fake.test/", tploader.WithHTTPClient(client))
		assert.Error(t, err)
	})

	t.Run("sets User-Agent header", func(t *testing.T) {
		t.Parallel()

		var gotUA string

		client := mockHTTPClient(func(req *http.Request) (*http.Response, error) {
			gotUA = req.Header.Get("User-Agent")

			return fakeResp(http.StatusOK, "ok"), nil
		})

		_, err := tploader.FetchContentFromURL(t.Context(), "http://fake.test/", tploader.WithHTTPClient(client))
		assert.NoError(t, err)
		assert.Contains(t, gotUA, "error-pages/")
	})

	t.Run("whitespace padded URL", func(t *testing.T) {
		t.Parallel()

		client := mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
			return fakeResp(http.StatusOK, "ok"), nil
		})

		got, err := tploader.FetchContentFromURL(t.Context(), "  http://fake.test/  ", tploader.WithHTTPClient(client))
		assert.NoError(t, err)
		assert.Equal(t, "ok", string(got))
	})
}

func TestReadContentFromFile(t *testing.T) {
	t.Parallel()

	t.Run("regular file", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.html")
		assert.NoError(t, err)
		_, err = io.WriteString(f, "template content")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		got, err := tploader.ReadContentFromFile(f.Name())
		assert.NoError(t, err)
		assert.Equal(t, "template content", string(got))
	})

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.html")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		got, err := tploader.ReadContentFromFile(f.Name())
		assert.NoError(t, err)
		assert.Equal(t, "", string(got))
	})

	t.Run("whitespace padded path", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.html")
		assert.NoError(t, err)
		_, err = io.WriteString(f, "hello")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		got, err := tploader.ReadContentFromFile("  " + f.Name() + "  ")
		assert.NoError(t, err)
		assert.Equal(t, "hello", string(got))
	})

	t.Run("nonexistent file", func(t *testing.T) {
		t.Parallel()

		_, err := tploader.ReadContentFromFile("/nonexistent/path/file.html")
		assert.Error(t, err)
	})

	t.Run("content exactly at size limit", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.txt")
		assert.NoError(t, err)
		_, err = io.WriteString(f, strings.Repeat("x", 100))
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		got, err := tploader.ReadContentFromFile(f.Name(), tploader.WithMaxTemplateSize(100))
		assert.NoError(t, err)
		assert.Equal(t, 100, len(got))
	})

	t.Run("content one byte over size limit", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.txt")
		assert.NoError(t, err)
		_, err = io.WriteString(f, strings.Repeat("x", 101))
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		_, err = tploader.ReadContentFromFile(f.Name(), tploader.WithMaxTemplateSize(100))
		assert.ErrorContains(t, err, "exceeds 100 bytes")
	})
}

func TestLoadTemplateContent(t *testing.T) {
	t.Parallel()

	t.Run("URL source", func(t *testing.T) {
		t.Parallel()

		client := mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
			return fakeResp(http.StatusOK, "template from URL"), nil
		})

		got, err := tploader.LoadTemplateContent(t.Context(), "http://fake.test/", tploader.WithHTTPClient(client))
		assert.NoError(t, err)
		assert.Equal(t, "template from URL", got)
	})

	t.Run("URL fetch error", func(t *testing.T) {
		t.Parallel()

		client := mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
			return fakeResp(http.StatusServiceUnavailable, ""), nil
		})

		_, err := tploader.LoadTemplateContent(t.Context(), "http://fake.test/", tploader.WithHTTPClient(client))
		assert.ErrorContains(t, err, "fetch template from URL")
	})

	t.Run("file path source", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.html")
		assert.NoError(t, err)
		_, err = io.WriteString(f, "template from file")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		got, err := tploader.LoadTemplateContent(t.Context(), f.Name())
		assert.NoError(t, err)
		assert.Equal(t, "template from file", got)
	})

	t.Run("literal template", func(t *testing.T) {
		t.Parallel()

		const src = `{{ .Code }} {{ .Message }}`

		got, err := tploader.LoadTemplateContent(t.Context(), src)
		assert.NoError(t, err)
		assert.Equal(t, src, got)
	})

	t.Run("URL empty body returns error", func(t *testing.T) {
		t.Parallel()

		client := mockHTTPClient(func(_ *http.Request) (*http.Response, error) {
			return fakeResp(http.StatusOK, ""), nil
		})

		_, err := tploader.LoadTemplateContent(t.Context(), "http://fake.test/", tploader.WithHTTPClient(client))
		assert.ErrorContains(t, err, "empty content from URL")
	})

	t.Run("empty file == error", func(t *testing.T) {
		t.Parallel()

		f, err := os.CreateTemp(t.TempDir(), "*.html")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		got, err := tploader.LoadTemplateContent(t.Context(), f.Name())
		assert.Empty(t, got)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "empty content from file")
	})
}
