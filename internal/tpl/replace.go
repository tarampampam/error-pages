package tpl

import "bytes"

type Replaces struct {
	Code        string
	Message     string
	Description string
}

const (
	tknCode byte = iota + 1
	tknMessage
	tknDescription
)

var tknSets = map[byte][][]byte{ //nolint:gochecknoglobals
	tknCode:        {[]byte("{{code}}"), []byte("{{ code }}")},
	tknMessage:     {[]byte("{{message}}"), []byte("{{ message }}")},
	tknDescription: {[]byte("{{description}}"), []byte("{{ description }}")},
}

// Replace found tokens in the incoming slice with passed tokens.
func Replace(in []byte, re Replaces) []byte {
	for tkn, set := range tknSets {
		var replaceWith []byte

		switch tkn {
		case tknCode:
			replaceWith = []byte(re.Code)
		case tknMessage:
			replaceWith = []byte(re.Message)
		case tknDescription:
			replaceWith = []byte(re.Description)
		default:
			panic("tpl: unsupported token")
		}

		if len(replaceWith) > 0 {
			for i := 0; i < len(set); i++ {
				in = bytes.ReplaceAll(in, set[i], replaceWith)
			}
		}
	}

	return in
}
