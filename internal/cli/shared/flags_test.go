package shared_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/config"
)

func TestListenAddrFlag(t *testing.T) {
	t.Parallel()

	var flag = shared.ListenAddrFlag

	assert.Equal(t, "listen", flag.Name)
	assert.Equal(t, "0.0.0.0", flag.Value)
	assert.Contains(t, flag.Sources.String(), "LISTEN_ADDR")

	for giveValue, wantErrMsg := range map[string]string{
		flag.Value: "", // default value

		// ipv4
		"0.0.0.0":         "",
		"127.0.0.1":       "",
		"255.255.255.255": "",

		// ipv6
		"::":  "",
		"::1": "",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334": "",
		"2001:db8:85a3:0:0:8a2e:370:7334":         "",
		"2001:db8:85a3::8a2e:370:7334":            "",
		"2001:db8::8a2e:370:7334":                 "",
		"2001:db8::7334":                          "",
		"2001:db8::":                              "",
		"2001:db8:0:0:1::1":                       "",
		"2001:db8:0:0:1::":                        "",

		// invalid
		"":                "missing IP address",
		"255.255.255.256": "wrong IP address [255.255.255.256] for listening",
		"example.com":     "wrong IP address [example.com] for listening",
		"123.123.abc.123": "wrong IP address [123.123.abc.123] for listening",
		"foo:123:321":     "wrong IP address [foo:123:321] for listening",
		"2001:db8:0:0:1:": "wrong IP address [2001:db8:0:0:1:] for listening",
	} {
		t.Run(fmt.Sprintf("%s: %s", giveValue, wantErrMsg), func(t *testing.T) {
			if err := flag.Validator(giveValue); wantErrMsg != "" {
				assert.ErrorContains(t, err, wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListenPortFlag(t *testing.T) {
	t.Parallel()

	var flag = shared.ListenPortFlag

	assert.Equal(t, "port", flag.Name)
	assert.Equal(t, uint(8080), flag.Value)
	assert.Contains(t, flag.Sources.String(), "LISTEN_PORT")

	for giveValue, wantErrMsg := range map[uint]string{
		flag.Value: "", // default value
		1:          "",
		8080:       "",
		65535:      "",

		0:     "wrong TCP port number [0]",
		65536: "wrong TCP port number [65536]",
	} {
		t.Run(fmt.Sprintf("%d: %s", giveValue, wantErrMsg), func(t *testing.T) {
			if err := flag.Validator(giveValue); wantErrMsg != "" {
				assert.ErrorContains(t, err, wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddTemplatesFlag(t *testing.T) {
	t.Parallel()

	var flag = shared.AddTemplatesFlag

	assert.Equal(t, "add-template", flag.Name)
	assert.Contains(t, flag.Sources.String(), "ADD_TEMPLATE")

	for wantErrMsg, giveValue := range map[string][]string{
		"missing template path":     {""},
		"wrong template path [.]":   {".", "./"},
		"wrong template path [..]":  {"..", "../"},
		"wrong template path [foo]": {"foo"},
		"":                          {"./flags.go"},
	} {
		t.Run(fmt.Sprintf("%s: %s", giveValue, wantErrMsg), func(t *testing.T) {
			if err := flag.Validator(giveValue); wantErrMsg != "" {
				assert.ErrorContains(t, err, wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDisableTemplateNamesFlag(t *testing.T) {
	t.Parallel()

	var flag = shared.DisableTemplateNamesFlag

	assert.Equal(t, "disable-template", flag.Name)
}

func TestAddHTTPCodesFlag(t *testing.T) {
	t.Parallel()

	var flag = shared.AddHTTPCodesFlag

	assert.Equal(t, "add-code", flag.Name)

	for name, tt := range map[string]struct {
		giveValue  map[string]string
		wantErrMsg string
	}{
		"common": {
			giveValue: map[string]string{
				"200": "foo/bar",
				"404": "foo",
				"2**": "baz",
			},
		},

		"missing HTTP code": {
			giveValue:  map[string]string{"": "foo/bar"},
			wantErrMsg: "missing HTTP code",
		},
		"wrong HTTP code [6]": {
			giveValue:  map[string]string{"6": "foo"},
			wantErrMsg: "wrong HTTP code [6]: it should be 3 characters long",
		},
		"wrong HTTP code [66]": {
			giveValue:  map[string]string{"66": "foo"},
			wantErrMsg: "wrong HTTP code [66]: it should be 3 characters long",
		},
		"wrong HTTP code [1000]": {
			giveValue:  map[string]string{"1000": "foo"},
			wantErrMsg: "wrong HTTP code [1000]: it should be 3 characters long",
		},
		"missing message and description": {
			giveValue:  map[string]string{"200": "//"},
			wantErrMsg: "wrong message/description format for HTTP code [200]: //",
		},
		"missing message": {
			giveValue:  map[string]string{"200": "/bar"},
			wantErrMsg: "missing message for HTTP code [200]",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := flag.Validator(tt.giveValue); tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseHTTPCodes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, shared.ParseHTTPCodes(nil), map[string]config.CodeDescription{})

	assert.Equal(t,
		shared.ParseHTTPCodes(map[string]string{"200": "msg"}),
		map[string]config.CodeDescription{"200": {Message: "msg", Description: ""}},
	)

	assert.Equal(t,
		shared.ParseHTTPCodes(map[string]string{"200": "/aaa"}),
		map[string]config.CodeDescription{"200": {Message: "", Description: "aaa"}},
	)

	assert.Equal(t, // not sure here
		shared.ParseHTTPCodes(map[string]string{"aa": "////aaa"}),
		map[string]config.CodeDescription{"aa": {Message: "", Description: "///aaa"}},
	)

	assert.Equal(t,
		shared.ParseHTTPCodes(map[string]string{"200": "msg/desc"}),
		map[string]config.CodeDescription{"200": {Message: "msg", Description: "desc"}},
	)

	assert.Equal(t,
		shared.ParseHTTPCodes(map[string]string{
			"200": "msg/desc",
			"foo": "Word word/Desc desc // adsadas",
		}),
		map[string]config.CodeDescription{
			"200": {Message: "msg", Description: "desc"},
			"foo": {Message: "Word word", Description: "Desc desc // adsadas"},
		},
	)
}

func TestDisableL10nFlag(t *testing.T) {
	t.Parallel()

	var flag = shared.DisableL10nFlag

	assert.Equal(t, "disable-l10n", flag.Name)
	assert.Contains(t, flag.Sources.String(), "DISABLE_L10N")
}

func TestDisableMinificationFlag(t *testing.T) {
	t.Parallel()

	var flag = shared.DisableMinificationFlag

	assert.Equal(t, "disable-minification", flag.Name)
	assert.Contains(t, flag.Sources.String(), "DISABLE_MINIFICATION")
}
