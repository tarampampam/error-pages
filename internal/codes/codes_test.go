package codes_test

import (
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/codes"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("not empty", func(t *testing.T) {
		t.Parallel()

		assertFalse(t, len(codes.New()) == 0)
	})

	t.Run("contains known codes", func(t *testing.T) {
		t.Parallel()

		c := codes.New()

		for code, wantShort := range map[string]string{
			"400": "Bad Request",
			"401": "Unauthorized",
			"403": "Forbidden",
			"404": "Not Found",
			"405": "Method Not Allowed",
			"407": "Proxy Authentication Required",
			"408": "Request Timeout",
			"409": "Conflict",
			"410": "Gone",
			"411": "Length Required",
			"412": "Precondition Failed",
			"413": "Payload Too Large",
			"416": "Requested Range Not Satisfiable",
			"418": "I'm a teapot",
			"429": "Too Many Requests",
			"500": "Internal Server Error",
			"502": "Bad Gateway",
			"503": "Service Unavailable",
			"504": "Gateway Timeout",
			"505": "HTTP Version Not Supported",
		} {
			t.Run(code, func(t *testing.T) {
				t.Parallel()

				desc, ok := c[code]
				assertTrue(t, ok)
				assertEqual(t, wantShort, desc.Short)
			})
		}
	})

	t.Run("returns independent clone", func(t *testing.T) {
		t.Parallel()

		c := codes.New()
		delete(c, "404")

		_, ok := codes.New()["404"]
		assertTrue(t, ok)
	})
}

func TestCodes_Find(t *testing.T) {
	t.Parallel()

	common := codes.Codes{
		"101": {Short: "Upgrade"},               // 101
		"1xx": {Short: "Informational"},         // 102-199
		"200": {Short: "OK"},                    // 200
		"20*": {Short: "Success"},               // 201-209
		"2**": {Short: "Success, but..."},       // 210-299
		"3**": {Short: "Redirection"},           // 300-399
		"404": {Short: "Not Found"},             // 404
		"405": {Short: "Method Not Allowed"},    // 405
		"500": {Short: "Internal Server Error"}, // 500
		"501": {Short: "Not Implemented"},       // 501
		"502": {Short: "Bad Gateway"},           // 502
		"503": {Short: "Service Unavailable"},   // 503
		"5XX": {Short: "Server Error"},          // 504-599
	}

	ladder := codes.Codes{
		"123": {Short: "Full triple"},
		"***": {Short: "Triple"},
		"12":  {Short: "Full double"},
		"**":  {Short: "Double"},
		"1":   {Short: "Full single"},
		"*":   {Short: "Single"},
	}

	for name, tt := range map[string]struct {
		giveCodes codes.Codes
		giveCode  uint16

		wantShort    string
		wantNotFound bool
	}{
		"101 - exact match":           {giveCodes: common, giveCode: 101, wantShort: "Upgrade"},
		"102 - multi-wildcard match":  {giveCodes: common, giveCode: 102, wantShort: "Informational"},
		"110 - multi-wildcard match":  {giveCodes: common, giveCode: 110, wantShort: "Informational"},
		"111 - multi-wildcard match":  {giveCodes: common, giveCode: 111, wantShort: "Informational"},
		"199 - multi-wildcard match":  {giveCodes: common, giveCode: 199, wantShort: "Informational"},
		"200 - exact match":           {giveCodes: common, giveCode: 200, wantShort: "OK"},
		"201 - single-wildcard match": {giveCodes: common, giveCode: 201, wantShort: "Success"},
		"209 - single-wildcard match": {giveCodes: common, giveCode: 209, wantShort: "Success"},
		"210 - multi-wildcard match":  {giveCodes: common, giveCode: 210, wantShort: "Success, but..."},
		"234 - multi-wildcard match":  {giveCodes: common, giveCode: 234, wantShort: "Success, but..."},
		"299 - multi-wildcard match":  {giveCodes: common, giveCode: 299, wantShort: "Success, but..."},
		"300 - multi-wildcard match":  {giveCodes: common, giveCode: 300, wantShort: "Redirection"},
		"301 - multi-wildcard match":  {giveCodes: common, giveCode: 301, wantShort: "Redirection"},
		"311 - multi-wildcard match":  {giveCodes: common, giveCode: 311, wantShort: "Redirection"},
		"399 - multi-wildcard match":  {giveCodes: common, giveCode: 399, wantShort: "Redirection"},
		"400 - not found":             {giveCodes: common, giveCode: 400, wantNotFound: true},
		"403 - not found":             {giveCodes: common, giveCode: 403, wantNotFound: true},
		"404 - exact match":           {giveCodes: common, giveCode: 404, wantShort: "Not Found"},
		"405 - exact match":           {giveCodes: common, giveCode: 405, wantShort: "Method Not Allowed"},
		"410 - not found":             {giveCodes: common, giveCode: 410, wantNotFound: true},
		"450 - not found":             {giveCodes: common, giveCode: 450, wantNotFound: true},
		"499 - not found":             {giveCodes: common, giveCode: 499, wantNotFound: true},
		"500 - exact match":           {giveCodes: common, giveCode: 500, wantShort: "Internal Server Error"},
		"501 - exact match":           {giveCodes: common, giveCode: 501, wantShort: "Not Implemented"},
		"502 - exact match":           {giveCodes: common, giveCode: 502, wantShort: "Bad Gateway"},
		"503 - exact match":           {giveCodes: common, giveCode: 503, wantShort: "Service Unavailable"},
		"504 - multi-wildcard match":  {giveCodes: common, giveCode: 504, wantShort: "Server Error"},
		"505 - multi-wildcard match":  {giveCodes: common, giveCode: 505, wantShort: "Server Error"},
		"599 - multi-wildcard match":  {giveCodes: common, giveCode: 599, wantShort: "Server Error"},
		"600 - not found":             {giveCodes: common, giveCode: 600, wantNotFound: true},

		"ladder - strict triple match": {giveCodes: ladder, giveCode: 123, wantShort: "Full triple"},
		"ladder - triple wildcard":     {giveCodes: ladder, giveCode: 321, wantShort: "Triple"},
		"ladder - strict double match": {giveCodes: ladder, giveCode: 12, wantShort: "Full double"},
		"ladder - double wildcard":     {giveCodes: ladder, giveCode: 21, wantShort: "Double"},
		"ladder - strict single match": {giveCodes: ladder, giveCode: 1, wantShort: "Full single"},
		"ladder - single wildcard":     {giveCodes: ladder, giveCode: 2, wantShort: "Single"},

		"empty map": {giveCodes: codes.Codes{}, giveCode: 404, wantNotFound: true},
		"zero code": {giveCodes: common, giveCode: 0, wantNotFound: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for range 10 { // repeat to ensure the function is idempotent
				desc, found := tt.giveCodes.Find(tt.giveCode)

				if !tt.wantNotFound {
					assertTrue(t, found)
					assertEqual(t, tt.wantShort, desc.Short)
				} else {
					assertFalse(t, found)
					assertEmpty(t, desc)
				}
			}
		})
	}
}

// --------------------------------------------------------------------------------------------------------------------

// assertTrue is a helper function that asserts that the given condition is true.
func assertTrue(t *testing.T, condition bool) {
	t.Helper()

	if !condition {
		t.Errorf("expected condition to be true, but it was false")
	}
}

// assertFalse is a helper function that asserts that the given condition is false.
func assertFalse(t *testing.T, condition bool) {
	t.Helper()

	if condition {
		t.Errorf("expected condition to be false, but it was true")
	}
}

// assertEqual is a helper function that asserts that the expected and actual values are equal.
func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

// assertEmpty is a helper function that asserts that the given value is the zero value of its type.
func assertEmpty[T comparable](t *testing.T, value T) {
	t.Helper()

	var zero T

	if value != zero {
		t.Errorf("expected empty value, got %v", value)
	}
}
