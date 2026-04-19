package codes

import (
	"maps"
	"slices"
	"strconv"
)

// Description holds HTTP error information.
type Description struct {
	// Short is a short description of the HTTP error.
	Short string

	// Full is a longer description of the HTTP error.
	Full string
}

// Codes is a map of HTTP codes to their descriptions.
//
// The codes may be written in a non-strict manner. For example, they may be "4xx", "4XX", or "4**".
// If the map contains both "404" and "4xx" keys, and we search via [Codes.Find] for "404", the "404" key will
// be returned.
// However, if we search for "405", "400", or any non-existing code that starts with "4" and its length is 3,
// the value under the key "4xx" will be retrieved.
//
// The length of the code (in string format) is matter.
type Codes map[string]Description

var builtInCodes = Codes{ //nolint:gochecknoglobals
	"400": {"Bad Request", "The server did not understand the request"},
	"401": {"Unauthorized", "The requested page needs a username and a password"},
	"403": {"Forbidden", "Access is forbidden to the requested page"},
	"404": {"Not Found", "The server can not find the requested page"},
	"405": {"Method Not Allowed", "The method specified in the request is not allowed"},
	"407": {"Proxy Authentication Required", "You must authenticate with a proxy server before this request can be served"}, //nolint:lll
	"408": {"Request Timeout", "The request took longer than the server was prepared to wait"},
	"409": {"Conflict", "The request could not be completed because of a conflict"},
	"410": {"Gone", "The requested page is no longer available"},
	"411": {"Length Required", "The \"Content-Length\" is not defined. The server will not accept the request without it"},
	"412": {"Precondition Failed", "The pre condition given in the request evaluated to false by the server"},
	"413": {"Payload Too Large", "The server will not accept the request, because the request entity is too large"},
	"416": {"Requested Range Not Satisfiable", "The requested byte range is not available and is out of bounds"},
	"418": {"I'm a teapot", "Attempt to brew coffee with a teapot is not supported"},
	"429": {"Too Many Requests", "Too many requests in a given amount of time"},
	"500": {"Internal Server Error", "The server met an unexpected condition"},
	"502": {"Bad Gateway", "The server received an invalid response from the upstream server"},
	"503": {"Service Unavailable", "The server is temporarily overloading or down"},
	"504": {"Gateway Timeout", "The gateway has timed out"},
	"505": {"HTTP Version Not Supported", "The server does not support the \"http protocol\" version"},
}

// New returns a new Codes map with the built-in HTTP codes and their descriptions.
func New() Codes { return maps.Clone(builtInCodes) }

// Find returns the description of the given HTTP code. If the code is not found, it returns false.
func (c Codes) Find(code uint16) (Description, bool) {
	if len(c) == 0 { // happiest path ;)
		return Description{}, false
	}

	// stack-allocated buffer, uint16 max = 65535 (5 digits)
	var buf [5]byte

	str := strconv.AppendUint(buf[:0], uint64(code), 10) //nolint:mnd

	if desc, ok := c[string(str)]; ok { // exact match
		return desc, true
	}

	var (
		bestKey string
		bestWC  = -1
	)

	for key := range c {
		if len(key) != len(str) || (!isWildcard(key[0]) && key[0] != str[0]) {
			continue // skip keys that are of different length or don't start with the same character or a wildcard
		}

		wc, matched := 0, true

		for i := range str {
			if kb := key[i]; isWildcard(kb) {
				wc++
			} else if kb != str[i] {
				matched = false

				break
			}
		}

		if matched && (bestWC < 0 || wc < bestWC) {
			bestKey, bestWC = key, wc
		}
	}

	if bestWC < 0 {
		return Description{}, false
	}

	return c[bestKey], true
}

func isWildcard(b byte) bool { return b == '*' || b == 'x' || b == 'X' }

// Codes returns all HTTP codes sorted alphabetically.
func (c Codes) Codes() []string { return slices.Sorted(maps.Keys(c)) }
