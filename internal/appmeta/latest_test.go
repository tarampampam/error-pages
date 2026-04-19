package appmeta_test

import (
	"net/http"
	"strings"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
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
				assertNoError(t, err)
				assertEqual(t, tt.wantVersion, latest)
			} else {
				assertErrorContains(t, err, tt.wantErrorSubstr)
				assertEqual(t, "", latest)
			}
		})
	}
}

// --------------------------------------------------------------------------------------------------------------------

// assertNoError is a helper function for tests that checks if the error is nil, and fails the test if it is not.
func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// assertErrorContains is a helper function for tests that checks if the error message contains the given substring,
// and fails the test if it does not.
func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q, but got nil", substr)
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("expected error containing %q, but got: %v", substr, err)
	}
}

// assertEqual is a helper function for tests that checks if the expected and actual values are equal, and fails the
// test if they are not.
func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Fatalf("expected: %v, got: %v", expected, actual)
	}
}
