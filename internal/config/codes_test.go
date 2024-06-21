package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/config"
)

func TestCodes_Common(t *testing.T) {
	t.Parallel()

	var codes = make(config.Codes)

	t.Run("initial state", func(t *testing.T) {
		require.Empty(t, codes.Codes())
		require.Empty(t, codes.Has("404"))

		var got, ok = codes.Get("404")

		require.Empty(t, got)
		require.False(t, ok)
	})

	t.Run("add a code", func(t *testing.T) {
		codes["404"] = config.CodeDescription{Message: "Not Found"}

		assert.True(t, codes.Has("404"))
		assert.Equal(t, []string{"404"}, codes.Codes())

		var got, ok = codes.Get("404")

		assert.Equal(t, got.Message, "Not Found")
		assert.True(t, ok)
	})
}

func TestCodes_Find(t *testing.T) {
	t.Parallel()

	//nolint:typecheck
	var common = config.Codes{
		"101": {Message: "Upgrade"},               // 101
		"1xx": {Message: "Informational"},         // 102-199
		"200": {Message: "OK"},                    // 200
		"20*": {Message: "Success"},               // 201-209
		"2**": {Message: "Success, but..."},       // 210-299
		"3**": {Message: "Redirection"},           // 300-399
		"404": {Message: "Not Found"},             // 404
		"405": {Message: "Method Not Allowed"},    // 405
		"500": {Message: "Internal Server Error"}, // 500
		"501": {Message: "Not Implemented"},       // 501
		"502": {Message: "Bad Gateway"},           // 502
		"503": {Message: "Service Unavailable"},   // 503
		"5XX": {Message: "Server Error"},          // 504-599
	}

	var ladder = config.Codes{
		"123": {Message: "Full triple"},
		"***": {Message: "Triple"},
		"12":  {Message: "Full double"},
		"**":  {Message: "Double"},
		"1":   {Message: "Full single"},
		"*":   {Message: "Single"},
	}

	for name, _tt := range map[string]struct {
		giveCodes config.Codes
		giveCode  uint16

		wantMessage  string
		wantNotFound bool
	}{
		"101 - exact match":           {giveCodes: common, giveCode: 101, wantMessage: "Upgrade"},
		"102 - multi-wildcard match":  {giveCodes: common, giveCode: 102, wantMessage: "Informational"},
		"110 - multi-wildcard match":  {giveCodes: common, giveCode: 110, wantMessage: "Informational"},
		"111 - multi-wildcard match":  {giveCodes: common, giveCode: 111, wantMessage: "Informational"},
		"199 - multi-wildcard match":  {giveCodes: common, giveCode: 199, wantMessage: "Informational"},
		"200 - exact match":           {giveCodes: common, giveCode: 200, wantMessage: "OK"},
		"201 - single-wildcard match": {giveCodes: common, giveCode: 201, wantMessage: "Success"},
		"209 - single-wildcard match": {giveCodes: common, giveCode: 209, wantMessage: "Success"},
		"210 - multi-wildcard match":  {giveCodes: common, giveCode: 210, wantMessage: "Success, but..."},
		"234 - multi-wildcard match":  {giveCodes: common, giveCode: 234, wantMessage: "Success, but..."},
		"299 - multi-wildcard match":  {giveCodes: common, giveCode: 299, wantMessage: "Success, but..."},
		"300 - multi-wildcard match":  {giveCodes: common, giveCode: 300, wantMessage: "Redirection"},
		"301 - multi-wildcard match":  {giveCodes: common, giveCode: 301, wantMessage: "Redirection"},
		"311 - multi-wildcard match":  {giveCodes: common, giveCode: 311, wantMessage: "Redirection"},
		"399 - multi-wildcard match":  {giveCodes: common, giveCode: 399, wantMessage: "Redirection"},
		"400 - not found":             {giveCodes: common, giveCode: 400, wantNotFound: true},
		"403 - not found":             {giveCodes: common, giveCode: 403, wantNotFound: true},
		"404 - exact match":           {giveCodes: common, giveCode: 404, wantMessage: "Not Found"},
		"405 - exact match":           {giveCodes: common, giveCode: 405, wantMessage: "Method Not Allowed"},
		"410 - not found":             {giveCodes: common, giveCode: 410, wantNotFound: true},
		"450 - not found":             {giveCodes: common, giveCode: 450, wantNotFound: true},
		"499 - not found":             {giveCodes: common, giveCode: 499, wantNotFound: true},
		"500 - exact match":           {giveCodes: common, giveCode: 500, wantMessage: "Internal Server Error"},
		"501 - exact match":           {giveCodes: common, giveCode: 501, wantMessage: "Not Implemented"},
		"502 - exact match":           {giveCodes: common, giveCode: 502, wantMessage: "Bad Gateway"},
		"503 - exact match":           {giveCodes: common, giveCode: 503, wantMessage: "Service Unavailable"},
		"504 - multi-wildcard match":  {giveCodes: common, giveCode: 504, wantMessage: "Server Error"},
		"505 - multi-wildcard match":  {giveCodes: common, giveCode: 505, wantMessage: "Server Error"},
		"599 - multi-wildcard match":  {giveCodes: common, giveCode: 599, wantMessage: "Server Error"},
		"600 - not found":             {giveCodes: common, giveCode: 600, wantNotFound: true},

		"ladder - strict triple match": {giveCodes: ladder, giveCode: 123, wantMessage: "Full triple"},
		"ladder - triple wildcard":     {giveCodes: ladder, giveCode: 321, wantMessage: "Triple"},
		"ladder - strict double match": {giveCodes: ladder, giveCode: 12, wantMessage: "Full double"},
		"ladder - double wildcard":     {giveCodes: ladder, giveCode: 21, wantMessage: "Double"},
		"ladder - strict single match": {giveCodes: ladder, giveCode: 1, wantMessage: "Full single"},
		"ladder - single wildcard":     {giveCodes: ladder, giveCode: 2, wantMessage: "Single"},

		"empty map": {giveCodes: config.Codes{}, giveCode: 404, wantNotFound: true},
		"zero code": {giveCodes: common, giveCode: 0, wantNotFound: true},
	} {
		var tt = _tt

		t.Run(name, func(t *testing.T) {
			for i := 0; i < 100; i++ { // repeat the test to ensure the function is idempotent
				var desc, found = tt.giveCodes.Find(tt.giveCode)

				if !tt.wantNotFound {
					require.Truef(t, found, "should have found something")
					require.Equal(t, tt.wantMessage, desc.Message)
				} else {
					require.Falsef(t, found, "should not have found anything, but got: %v", desc)
					require.Empty(t, desc)
				}
			}
		})
	}
}
