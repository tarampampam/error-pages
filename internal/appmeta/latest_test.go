package appmeta_test

import (
	"net/http"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestLatest(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		giveStatusCode int
		giveLocation   string

		wantVersion     string
		wantErrorSubstr string
	}{
		"success": {
			giveStatusCode: http.StatusFound,
			giveLocation:   "https://github.com/tarampampam/webhook-tester/releases/tag/V1.2.0/foo/bar?baz=qux#quux",
			wantVersion:    "1.2.0",
		},
		"success without v prefix": {
			giveStatusCode: http.StatusFound,
			giveLocation:   "https://github.com/tarampampam/webhook-tester/releases/tag/1.2.0/foo/bar?baz=qux#quux",
			wantVersion:    "1.2.0",
		},

		"unexpected status code": {
			giveStatusCode:  http.StatusNotFound,
			wantErrorSubstr: "unexpected status code: 404",
		},
		"redirect location is malformed": {
			giveStatusCode:  http.StatusFound,
			giveLocation:    "qwe",
			wantErrorSubstr: "unexpected location path: qwe",
		},
		"too short location link": {
			giveStatusCode:  http.StatusFound,
			giveLocation:    "https://github.com/owner/repo/foo",
			wantErrorSubstr: "unexpected location path: /owner/repo/foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var client httpClientFunc = func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tt.giveStatusCode,
					Header:     http.Header{"Location": []string{tt.giveLocation}},
				}, nil
			}

			latest, err := appmeta.Latest(t.Context(), client)

			if tt.wantErrorSubstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVersion, latest)
			} else {
				assert.ErrorContains(t, err, tt.wantErrorSubstr)
				assert.Equal(t, "", latest)
			}
		})
	}
}
