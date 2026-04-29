package jsmin_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
	"gh.tarampamp.am/error-pages/v4/l10n/generate/jsmin"
)

func TestMinifyString(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		in   string
		want string
	}{
		"empty": {in: "", want: ""},
		"simple assignment": {
			in:   "var x = 1;",
			want: "var x=1;",
		},
		// A '\n' following ';' is dropped by jsmin — ';' is not in
		// the set of characters that allow a trailing newline.
		"line comment stripped": {
			in:   "var x = 1; // trailing comment\nvar y = 2;",
			want: "var x=1;var y=2;",
		},
		"block comment stripped": {
			in:   "var /* inline */ x = 1;",
			want: "var x=1;",
		},
		"string literal preserved": {
			in:   `var s = "hello  world"; // comment`,
			want: "var s=\"hello  world\";",
		},
		"string with escaped quote": {
			in:   `var s = "a\"b";`,
			want: "var s=\"a\\\"b\";",
		},
		"template literal preserved": {
			in:   "var s = `   spaces   `;",
			want: "var s=`   spaces   `;",
		},
		"regex literal preserved": {
			in:   "var re = /ab+c/i;",
			want: "var re=/ab+c/i;",
		},
		"regex with character class": {
			in:   "var re = /[a/b]+/;",
			want: "var re=/[a/b]+/;",
		},
		"division not mistaken for regex": {
			in:   "var x = a / b / c;",
			want: "var x=a/b/c;",
		},
		"ambiguous operator preserved": {
			in:   "var x = a + +b;",
			want: "var x=a+ +b;",
		},
		"utf8 identifiers preserved": {
			in:   "var π = 3.14; var λ = 2;",
			want: "var π=3.14;var λ=2;",
		},
		"utf8 BOM stripped": {
			in:   "\xEF\xBB\xBFvar x = 1;",
			want: "var x=1;",
		},
		// CR is normalised to LF; the resulting LF after ';' is
		// then dropped per the semicolon rule above.
		"CR translated to LF and dropped after semicolon": {
			in:   "var x = 1;\r\nvar y = 2;",
			want: "var x=1;var y=2;",
		},
		// The complementary case: a newline after '}' IS preserved.
		"newline preserved after close brace": {
			in:   "function f(){}\nvar y = 2;",
			want: "function f(){}\nvar y=2;",
		},
		"tabs become spaces": {
			in:   "var\tx\t=\t1;",
			want: "var x=1;",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := jsmin.MinifyString(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestCrockfordIsExample uses the canonical example from Crockford's JSMin README to catch regressions in the
// main-loop transitions.
func TestCrockfordIsExample(t *testing.T) {
	t.Parallel()

	const in = `var is = {
    ie:      navigator.appName == 'Microsoft Internet Explorer',
    java:    navigator.javaEnabled(),
    ns:      navigator.appName == 'Netscape',
    ua:      navigator.userAgent.toLowerCase(),
    version: parseFloat(navigator.appVersion.substr(21)) ||
             parseFloat(navigator.appVersion),
    win:     navigator.platform == 'Win32'
}
is.mac = is.ua.indexOf('mac') >= 0;
if (is.ua.indexOf('opera') >= 0) {
    is.ie = is.ns = false;
    is.opera = true;
}
if (is.ua.indexOf('gecko') >= 0) {
    is.ie = is.ns = false;
    is.gecko = true;
}
`

	// expected output per Crockford's reference implementation.
	const want = "var is={ie:navigator.appName=='Microsoft Internet Explorer'," +
		"java:navigator.javaEnabled()," +
		"ns:navigator.appName=='Netscape'," +
		"ua:navigator.userAgent.toLowerCase()," +
		"version:parseFloat(navigator.appVersion.substr(21))||parseFloat(navigator.appVersion)," +
		"win:navigator.platform=='Win32'}\n" +
		"is.mac=is.ua.indexOf('mac')>=0;if(is.ua.indexOf('opera')>=0){is.ie=is.ns=false;is.opera=true;}\n" +
		"if(is.ua.indexOf('gecko')>=0){is.ie=is.ns=false;is.gecko=true;}"

	got, err := jsmin.MinifyString(in)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMinifyErrors(t *testing.T) {
	t.Parallel()

	for name, tt := range map[string]struct {
		in      string
		wantErr error
	}{
		"unterminated block comment": {
			in:      "var x = 1; /* no closing",
			wantErr: jsmin.ErrUnterminatedComment,
		},
		"unterminated double-quoted string": {
			in:      `var x = "open`,
			wantErr: jsmin.ErrUnterminatedString,
		},
		"unterminated single-quoted string": {
			in:      "var x = 'open",
			wantErr: jsmin.ErrUnterminatedString,
		},
		"unterminated template literal": {
			in:      "var x = `open",
			wantErr: jsmin.ErrUnterminatedString,
		},
		"unterminated regex literal": {
			in:      "var x = (/abc",
			wantErr: jsmin.ErrUnterminatedRegex,
		},
		"unterminated character class in regex": {
			in:      "var x = (/[abc",
			wantErr: jsmin.ErrUnterminatedRegexSet,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := jsmin.MinifyString(tt.in)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestMinifyDoesNotSwallowReadErrors verifies that I/O errors from the
// source reader surface as errors from Minify rather than being treated
// as end-of-input.
func TestMinifyDoesNotSwallowReadErrors(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	var buf bytes.Buffer
	assert.ErrorIs(t, jsmin.Minify(&buf, &errorReader{err: boom}), boom)
}

// TestMinifyAcceptsStreamingReader ensures that non-byte readers are
// wrapped correctly (i.e. we don't silently depend on bufio being used).
func TestMinifyAcceptsStreamingReader(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	assert.NoError(t, jsmin.Minify(&buf, &byteByByteReader{s: "var x = 1;"}))
	assert.Equal(t, "var x=1;", buf.String())
}

type errorReader struct{ err error }

func (r *errorReader) Read(_ []byte) (int, error) { return 0, r.err }

// byteByByteReader returns one byte per Read call and does not implement
// io.ByteReader, forcing Minify's wrapping path to run.
type byteByByteReader struct {
	s string
	i int
}

func (r *byteByByteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}

	if len(p) == 0 {
		return 0, nil
	}

	p[0] = r.s[r.i]
	r.i++

	return 1, nil
}

func BenchmarkMinifyString(b *testing.B) {
	const src = `
/* A moderately sized chunk of javascript to exercise
   the hot path of the minifier. */
function fib(n) {
    if (n < 2) return n;        // base case
    var a = 0, b = 1, t;
    for (var i = 2; i <= n; i++) {
        t = a + b;
        a = b;
        b = t;
    }
    return b;
}
var re = /^[a-zA-Z_][a-zA-Z0-9_]*$/;
console.log(fib(30), re.test("hello"));
`

	b.ReportAllocs()
	b.SetBytes(int64(len(src)))

	for range b.N {
		if _, err := jsmin.MinifyString(src); err != nil {
			b.Fatal(err)
		}
	}
}
