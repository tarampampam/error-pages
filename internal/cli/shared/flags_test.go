package shared_test

import (
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/cli/shared"
	"gh.tarampamp.am/error-pages/v4/internal/codes"
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
