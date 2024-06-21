package shared_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
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
	assert.Equal(t, uint64(8080), flag.Value)
	assert.Contains(t, flag.Sources.String(), "LISTEN_PORT")

	for giveValue, wantErrMsg := range map[uint64]string{
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
