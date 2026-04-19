// Package templates contains the HTML templates for the application. The templates are embedded in the binary using
// the go:embed directive. This file is used to generate the embed.go file, which contains the embedded templates.
package templates

//go:generate go run ./generate/embed.go
