// Package templates contains the HTML/SJON/XML/etc templates for the app. The templates are embedded in the binary
// using the go:embed directive. This file is also used to generate the embed_html.go file, which contains the
// embedded HTML templates.
package templates

import _ "embed"

//go:generate go run ./generate/embed_html.go -src ./html -out ./embed_html.go

// JSON holds the embedded JSON template for error responses.
//
//go:embed default.tpl.json
var JSON string

// XML holds the embedded XML template for error responses.
//
//go:embed default.tpl.xml
var XML string

// PlaintText holds the embedded plain text template for error responses.
//
//go:embed default.tpl.txt
var PlaintText string
