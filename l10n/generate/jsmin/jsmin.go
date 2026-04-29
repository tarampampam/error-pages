// Package jsmin is a Go port of Douglas Crockford's JSMin (https://github.com/douglascrockford/JSMin). It minifies
// JavaScript source code by removing comments and collapsing insignificant whitespace while preserving string
// literals and regular-expression literals.
//
// The algorithm is byte-oriented. Input must be ASCII or UTF-8; other encodings are not supported. JSMin is a
// one-way transformation - the minified output cannot be mechanically restored to the original.
package jsmin

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Lexical errors returned when the input contains unterminated constructs.
var (
	ErrUnterminatedComment  = errors.New("jsmin: unterminated comment")
	ErrUnterminatedString   = errors.New("jsmin: unterminated string literal")
	ErrUnterminatedRegex    = errors.New("jsmin: unterminated regular expression literal")
	ErrUnterminatedRegexSet = errors.New("jsmin: unterminated set in regular expression literal")
)

const eof = -1

// Action determinations taken by the main loop.
const (
	actionOutput = 1 // emit A, promote B to A, pull the next B
	actionCopy   = 2 // discard A, promote B to A, pull the next B
	actionSkip   = 3 // just pull the next B (discarding B)
)

// Minify reads JavaScript source from src, minifies it, and writes the result to dst.
//
// If src does not already implement [io.ByteReader] it is wrapped in a [bufio.Reader]; likewise dst is wrapped in a
// [bufio.Writer] and flushed before return if it does not already implement [io.ByteWriter].
func Minify(dst io.Writer, src io.Reader) error {
	m := minifier{
		in: asByteReader(src),
		la: eof,
		x:  eof,
		y:  eof,
	}
	bw, flush := asByteWriter(dst)
	m.out = bw

	if err := m.run(); err != nil {
		return err
	}

	return flush()
}

// MinifyString is a convenience wrapper around Minify for in-memory input.
func MinifyString(src string) (string, error) {
	var b strings.Builder
	b.Grow(len(src)) // output is never larger than input.

	if err := Minify(&b, strings.NewReader(src)); err != nil {
		return "", err
	}

	return b.String(), nil
}

func asByteReader(r io.Reader) io.ByteReader {
	if br, ok := r.(io.ByteReader); ok {
		return br
	}

	return bufio.NewReader(r)
}

func asByteWriter(w io.Writer) (io.ByteWriter, func() error) {
	if bw, ok := w.(io.ByteWriter); ok {
		// caller owns the buffering decision; nothing to flush
		return bw, func() error { return nil }
	}

	bw := bufio.NewWriter(w)

	return bw, bw.Flush
}

// minifier holds the state machine's lookahead window and I/O endpoints.
type minifier struct {
	in  io.ByteReader
	out io.ByteWriter

	// a and b are the two-byte output/lookahead window the state machine operates on. After get()/next() returns,
	// `a` is the current character being considered and `b` is the upcoming one
	a, b int

	// x and y are the two most recent characters returned by next(), used to preserve whitespace between ambiguous
	// operator pairs (e.g. `+ +`)
	x, y int

	// la is single-byte lookahead buffer used by peek()
	la int

	suppressFirst bool
}

// bail is the panic payload used to unwind out of deeply nested parsing loops. It is recovered in run(), which
// turns it back into an error.
type bail struct{ err error }

func (m *minifier) fail(err error) { panic(bail{err}) }

// run invokes jsmin and converts internal bails into a returned error. Any other panic is re-raised verbatim.
func (m *minifier) run() (err error) {
	defer func() {
		switch r := recover().(type) {
		case nil:
		case bail:
			err = r.err
		default:
			panic(r)
		}
	}()

	m.suppressFirst = true
	m.jsmin()

	return nil
}

// put writes a single byte to the output. Write errors are surfaced via bail so callers of the state machine don't
// need to propagate them.
func (m *minifier) put(c int) {
	if m.suppressFirst {
		m.suppressFirst = false

		if c == '\n' {
			return
		}
	}

	if err := m.out.WriteByte(byte(c)); err != nil { //nolint:gosec
		m.fail(fmt.Errorf("jsmin: write error: %w", err))
	}
}

// get returns the next byte from the input, translating control characters to spaces (or '\r' to '\n'). Returns eof
// at end of input.
func (m *minifier) get() int {
	c := m.la

	m.la = eof
	if c == eof {
		b, err := m.in.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return eof
			}

			m.fail(fmt.Errorf("jsmin: read error: %w", err))
		}

		c = int(b)
	}

	// at this point c is a concrete byte value (0..255)
	switch {
	case c >= ' ', c == '\n':
		return c
	case c == '\r':
		return '\n'
	default:
		return ' '
	}
}

// peek returns the next byte without consuming it.
func (m *minifier) peek() int {
	m.la = m.get()

	return m.la
}

// next returns the next significant byte, transparently skipping over both `//` line comments and `/*` block comments.
func (m *minifier) next() int {
	c := m.get()
	if c == '/' {
		switch m.peek() {
		case '/':
			for {
				c = m.get()
				if c <= '\n' { // matches '\n' or eof
					break
				}
			}
		case '*':
			m.get() // consume the '*'

			for c != ' ' {
				switch m.get() {
				case '*':
					if m.peek() == '/' {
						m.get()

						c = ' '
					}
				case eof:
					m.fail(ErrUnterminatedComment)
				}
			}
		}
	}

	m.y, m.x = m.x, c

	return c
}

// action performs one of the three determinations above. The structure mirrors Crockford's jsmin.c action() but
// unrolls its implicit case fall-through into explicit sequential steps.
func (m *minifier) action(kind int) {
	// Step 1 - emit A and, if we're looking at an ambiguous operator pair (like `+ +`), preserve the separator.
	// Only for actionOutput.
	if kind == actionOutput {
		m.put(m.a)

		if (m.y == '\n' || m.y == ' ') && isAmbigOp(m.a) && isAmbigOp(m.b) {
			m.put(m.y)
		}
	}

	// Step 2 - advance A and, if it's a string-literal quote, copy the entire literal to the output. For actionOutput
	// and actionCopy.
	if kind == actionOutput || kind == actionCopy {
		m.a = m.b
		if m.a == '\'' || m.a == '"' || m.a == '`' {
			quote := m.a
			for {
				m.put(m.a)

				m.a = m.get()
				if m.a == quote {
					break
				}

				if m.a == '\\' {
					m.put(m.a)
					m.a = m.get()
				}

				if m.a == eof {
					m.fail(ErrUnterminatedString)
				}
			}
		}
	}

	// Step 3 - pull the next B. If it looks like the start of a regex literal (preceding token allows one), emit the
	// whole regex body.
	m.b = m.next()
	if m.b == '/' && startsRegex(m.a) {
		m.emitRegex()
	}
}

// emitRegex outputs the regex literal currently being opened (A holds the preceding token, B is the opening '/'). On
// successful exit, A is the closing '/' of the regex (already consumed but not yet emitted by this function - the
// outer loop emits it as a normal byte) and B is reset via next().
func (m *minifier) emitRegex() {
	m.put(m.a)
	// if A is itself / or * we must insert a space so the new regex opener is not mistaken for a continuation of
	// the previous token
	if m.a == '/' || m.a == '*' {
		m.put(' ')
	}

	m.put(m.b)

body:
	for {
		m.a = m.get()
		switch m.a {
		case '[':
			// character class: copy verbatim until unescaped ']'
			for {
				m.put(m.a)

				m.a = m.get()
				if m.a == ']' {
					break
				}

				if m.a == '\\' {
					m.put(m.a)
					m.a = m.get()
				}

				if m.a == eof {
					m.fail(ErrUnterminatedRegexSet)
				}
			}
			// fall through to emit the ']' below
		case '/':
			// Closing slash. A regex may not be immediately followed by // or /* because that would start a comment inside
			// what jsmin has to treat as a single token.
			//
			// NOTE: the original jsmin.c reports this as "unterminated set in Regular Expression literal" even though it is
			// not a character-class error. Behavior preserved for parity.
			if p := m.peek(); p == '/' || p == '*' {
				m.fail(ErrUnterminatedRegexSet)
			}

			break body
		case '\\':
			m.put(m.a)
			m.a = m.get()
			// Fall through to emit the escaped byte below.
		}

		if m.a == eof {
			m.fail(ErrUnterminatedRegex)
		}

		m.put(m.a)
	}

	m.b = m.next()
}

// jsmin is the main state-machine driver. It copies input to output, deleting characters that are insignificant to
// JavaScript: comments are removed, runs of whitespace are collapsed, tabs become spaces, and CRs become LFs.
func (m *minifier) jsmin() {
	// skip UTF-8 BOM if present.
	if m.peek() == 0xEF { //nolint:mnd
		m.get()
		m.get()
		m.get()
	}

	m.a = '\n'
	m.action(actionSkip)

	for m.a != eof {
		switch m.a {
		case ' ':
			if isAlphanum(m.b) {
				m.action(actionOutput)
			} else {
				m.action(actionCopy)
			}
		case '\n':
			switch m.b {
			case '{', '[', '(', '+', '-', '!', '~':
				m.action(actionOutput)
			case ' ':
				m.action(actionSkip)
			default:
				if isAlphanum(m.b) {
					m.action(actionOutput)
				} else {
					m.action(actionCopy)
				}
			}
		default:
			switch m.b {
			case ' ':
				if isAlphanum(m.a) {
					m.action(actionOutput)
				} else {
					m.action(actionSkip)
				}
			case '\n':
				switch m.a {
				case '}', ']', ')', '+', '-', '"', '\'', '`':
					m.action(actionOutput)
				default:
					if isAlphanum(m.a) {
						m.action(actionOutput)
					} else {
						m.action(actionSkip)
					}
				}
			default:
				m.action(actionOutput)
			}
		}
	}
}

// isAlphanum reports whether c should be treated as "identifier-ish": letter, digit, underscore, dollar, backslash,
// or any non-ASCII byte (which makes the minifier UTF-8 safe without decoding code points).
func isAlphanum(c int) bool {
	switch {
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	case c >= 'A' && c <= 'Z':
		return true
	case c == '_', c == '$', c == '\\':
		return true
	case c > 126: //nolint:mnd
		return true
	}

	return false
}

// isAmbigOp reports whether c is one of the operator characters whose adjacency would be ambiguous without an
// intervening space.
func isAmbigOp(c int) bool {
	switch c {
	case '+', '-', '*', '/':
		return true
	}

	return false
}

// startsRegex reports whether a '/' following c should be parsed as the start of a regex literal rather than
// as division.
func startsRegex(c int) bool {
	switch c {
	case '(', ',', '=', ':', '[', '!', '&', '|', '?', '+', '-', '~', '*', '/', '{', '}', ';':
		return true
	}

	return false
}
