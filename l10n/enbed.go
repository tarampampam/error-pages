package l10n

import _ "embed"

//go:embed l10n.js
var content string

// L10n returns the content of the JS file with a script for automatic error page localization.
func L10n() string { return content }
