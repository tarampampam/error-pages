package l10n

import _ "embed"

//go:generate go run ./generate/localize.go -locales ./locales.json -out ./localize.js -out-min localize.min.js

//go:embed localize.min.js
var content string

// L10n returns the content of the JS file with a script for automatic error page localization.
func L10n() string { return content }
