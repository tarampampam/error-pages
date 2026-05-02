package formats

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html"
)

// Format is an enumeration of supported content formats for error page responses.
type Format byte

const (
	PlainTextFormat Format = iota // default, plain text
	JSONFormat                    // json
	XMLFormat                     // xml
	HTMLFormat                    // html
)

// ContentType returns the MIME content type string for the format.
func (f Format) ContentType() string {
	switch f {
	case PlainTextFormat:
		return "text/plain; charset=utf-8"
	case HTMLFormat:
		return "text/html; charset=utf-8"
	case JSONFormat:
		return "application/json; charset=utf-8"
	case XMLFormat:
		return "application/xml; charset=utf-8"
	}

	return ""
}

// FormatError formats the given error message according to the format.
func (f Format) FormatError(errStr string) []byte {
	switch f {
	case JSONFormat:
		b, _ := json.Marshal(struct { //nolint:errcheck,errchkjson
			Error string `json:"error"`
		}{errStr})

		return b
	case XMLFormat:
		var buf bytes.Buffer
		buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<error>")
		_ = xml.EscapeText(&buf, []byte(errStr)) //nolint:errcheck
		buf.WriteString("</error>")

		return buf.Bytes()
	case HTMLFormat:
		return []byte("<!DOCTYPE html>\n" +
			"<html><head><meta charset=\"UTF-8\"></head><body>\n" + html.EscapeString(errStr) + "\n</body></html>")
	case PlainTextFormat:
		return []byte(errStr)
	}

	return []byte(errStr)
}
