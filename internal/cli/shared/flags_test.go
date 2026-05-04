package shared_test

import (
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/cli/shared"
	"gh.tarampamp.am/error-pages/v4/internal/codes"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNewDisableBuiltInCodesFlag(t *testing.T) {
	t.Parallel()

	f := shared.NewDisableBuiltInCodesFlag()

	assert.Equal(t, 1, len(f.Names))
	assert.Equal(t, "disable-built-in-codes", f.Names[0])
	assert.Equal(t, 1, len(f.EnvVars))
	assert.Equal(t, "DISABLE_BUILT_IN_CODES", f.EnvVars[0])
	assert.Equal(t, false, f.Default)
}

func TestNewAddHTTPCodesFlag(t *testing.T) {
	t.Parallel()

	f := shared.NewAddHTTPCodesFlag()

	assert.Equal(t, 1, len(f.Names))
	assert.Equal(t, "add-code", f.Names[0])
	assert.Equal(t, 1, len(f.EnvVars))
	assert.Equal(t, "ADD_CODE", f.EnvVars[0])
	assert.True(t, f.Validator != nil)

	assert.NoError(t, f.Validator(nil, "404=Not Found"))
	assert.Error(t, f.Validator(nil, "bad-entry"))
}

func TestNewHomepageURLFlag(t *testing.T) {
	t.Parallel()

	t.Run("names and env vars", func(t *testing.T) {
		t.Parallel()

		f := shared.NewHomepageURLFlag("")

		assert.Equal(t, 1, len(f.Names))
		assert.Equal(t, "homepage-url", f.Names[0])
		assert.Equal(t, 1, len(f.EnvVars))
		assert.Equal(t, "HOMEPAGE_URL", f.EnvVars[0])
		assert.True(t, f.Validator == nil)
	})

	t.Run("default is forwarded", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "", shared.NewHomepageURLFlag("").Default)
		assert.Equal(t, "/", shared.NewHomepageURLFlag("/").Default)
		assert.Equal(t, "https://app.example.com/home", shared.NewHomepageURLFlag("https://app.example.com/home").Default)
	})
}

func TestNewDisableL10nFlag(t *testing.T) {
	t.Parallel()

	f := shared.NewDisableL10nFlag()

	assert.Equal(t, 1, len(f.Names))
	assert.Equal(t, "disable-l10n", f.Names[0])
	assert.Equal(t, 1, len(f.EnvVars))
	assert.Equal(t, "DISABLE_L10N", f.EnvVars[0])
	assert.Equal(t, false, f.Default)
}

func TestParseAddHTTPCodes(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		give     string
		want     map[string]codes.Description
		checkErr func(*testing.T, error)
	}{
		"empty string": {
			give: "",
			want: map[string]codes.Description{},
		},
		"single entry/message only": {
			give: "404=Not Found",
			want: map[string]codes.Description{"404": {Short: "Not Found"}},
		},
		"single entry/message and description": {
			give: "500=Internal Server Error|Something went wrong on the server",
			want: map[string]codes.Description{
				"500": {Short: "Internal Server Error", Full: "Something went wrong on the server"},
			},
		},
		"multiple entries/double pipe separator": {
			give: "404=Not Found||500=Internal Server Error",
			want: map[string]codes.Description{
				"404": {Short: "Not Found"},
				"500": {Short: "Internal Server Error"},
			},
		},
		"multiple entries/newline separator": {
			give: "404=Not Found\n500=Internal Server Error",
			want: map[string]codes.Description{
				"404": {Short: "Not Found"},
				"500": {Short: "Internal Server Error"},
			},
		},
		"multiple entries/tab separator": {
			give: "404=Not Found\t500=Internal Server Error",
			want: map[string]codes.Description{
				"404": {Short: "Not Found"},
				"500": {Short: "Internal Server Error"},
			},
		},
		"empty entries in the middle are skipped": {
			give: "404=Not Found||||500=Internal Server Error",
			want: map[string]codes.Description{
				"404": {Short: "Not Found"},
				"500": {Short: "Internal Server Error"},
			},
		},
		"whitespace trimmed around code and message": {
			give: "  404  =  Not Found  ",
			want: map[string]codes.Description{"404": {Short: "Not Found"}},
		},
		"whitespace trimmed around description": {
			give: "404=Not Found|  The page was not found  ",
			want: map[string]codes.Description{"404": {Short: "Not Found", Full: "The page was not found"}},
		},
		"wildcard/star": {
			give: "4**=Client Error",
			want: map[string]codes.Description{"4**": {Short: "Client Error"}},
		},
		"wildcard/lowercase x": {
			give: "4xx=Client Error",
			want: map[string]codes.Description{"4xx": {Short: "Client Error"}},
		},
		"wildcard/uppercase X": {
			give: "4XX=Client Error",
			want: map[string]codes.Description{"4XX": {Short: "Client Error"}},
		},
		"override same code": {
			give: "404=First||404=Second",
			want: map[string]codes.Description{"404": {Short: "Second"}},
		},
		"missing equals sign": {
			give:     "404NotFound",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing '='") },
		},
		"empty code": {
			give:     "=Not Found",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing HTTP code") },
		},
		"code too short": {
			give:     "40=Not Found",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "must be 3 characters long") },
		},
		"code too long": {
			give:     "4044=Not Found",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "must be 3 characters long") },
		},
		"invalid character in code": {
			give: "40a=Not Found",
			checkErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "allowed characters are digits and wildcards")
			},
		},
		"empty message": {
			give:     "404=",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing message") },
		},
		"whitespace-only message": {
			give:     "404=   ",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing message") },
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := shared.ParseAddHTTPCodes(tt.give)

			if tt.checkErr != nil {
				tt.checkErr(t, err)
			} else {
				assert.NoError(t, err)
				assert.DeepEqual(t, tt.want, got)
			}
		})
	}
}

func TestNewAddLinksFlag(t *testing.T) {
	t.Parallel()

	f := shared.NewAddLinksFlag()

	assert.Equal(t, 1, len(f.Names))
	assert.Equal(t, "add-link", f.Names[0])
	assert.Equal(t, 1, len(f.EnvVars))
	assert.Equal(t, "ADD_LINK", f.EnvVars[0])
	assert.True(t, f.Validator != nil)

	assert.NoError(t, f.Validator(nil, "Status Page=https://status.example.com"))
	assert.Error(t, f.Validator(nil, "bad-entry"))
}

func TestParseLinks(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		give     string
		want     []tpl.Link
		checkErr func(*testing.T, error)
	}{
		"empty string": {
			give: "",
			want: []tpl.Link{},
		},
		"single entry": {
			give: "Status Page=https://status.example.com",
			want: []tpl.Link{{Label: "Status Page", URL: "https://status.example.com"}},
		},
		"url with equals sign": {
			give: "Search=https://example.com/search?q=foo&page=1",
			want: []tpl.Link{{Label: "Search", URL: "https://example.com/search?q=foo&page=1"}},
		},
		"multiple entries/double pipe separator": {
			give: "Status=https://status.example.com||Contact=https://example.com/contact",
			want: []tpl.Link{
				{Label: "Status", URL: "https://status.example.com"},
				{Label: "Contact", URL: "https://example.com/contact"},
			},
		},
		"multiple entries/newline separator": {
			give: "Status=https://status.example.com\nContact=https://example.com/contact",
			want: []tpl.Link{
				{Label: "Status", URL: "https://status.example.com"},
				{Label: "Contact", URL: "https://example.com/contact"},
			},
		},
		"multiple entries/tab separator": {
			give: "Status=https://status.example.com\tContact=https://example.com/contact",
			want: []tpl.Link{
				{Label: "Status", URL: "https://status.example.com"},
				{Label: "Contact", URL: "https://example.com/contact"},
			},
		},
		"empty entries in the middle are skipped": {
			give: "Status=https://status.example.com||||Contact=https://example.com/contact",
			want: []tpl.Link{
				{Label: "Status", URL: "https://status.example.com"},
				{Label: "Contact", URL: "https://example.com/contact"},
			},
		},
		"whitespace trimmed around label and url": {
			give: "  Status Page  =  https://status.example.com  ",
			want: []tpl.Link{{Label: "Status Page", URL: "https://status.example.com"}},
		},
		"missing equals sign": {
			give:     "Status Page",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing '='") },
		},
		"empty label": {
			give:     "=https://status.example.com",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing label") },
		},
		"whitespace-only label": {
			give:     "   =https://status.example.com",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing label") },
		},
		"empty url": {
			give:     "Status=",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing URL") },
		},
		"whitespace-only url": {
			give:     "Status=   ",
			checkErr: func(t *testing.T, err error) { assert.ErrorContains(t, err, "missing URL") },
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := shared.ParseLinks(tt.give)

			if tt.checkErr != nil {
				tt.checkErr(t, err)
			} else {
				assert.NoError(t, err)
				assert.DeepEqual(t, tt.want, got)
			}
		})
	}
}
