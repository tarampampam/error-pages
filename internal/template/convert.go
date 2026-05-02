package tpl

import (
	"regexp"
	"strings"
)

// v3tov4Tokens maps old function-style tokens to new dot-field paths.
//
// Deprecated: Temporary workaround, used by [convertV3toV4].
var v3tov4Tokens = map[string]string{ //nolint:gochecknoglobals
	"code":          ".StatusCode",
	"message":       ".Message",
	"description":   ".Description",
	"original_uri":  ".OriginalURI",
	"namespace":     ".Namespace",
	"ingress_name":  ".IngressName",
	"service_name":  ".ServiceName",
	"service_port":  ".ServicePort",
	"request_id":    ".RequestID",
	"forwarded_for": ".ForwardedFor",
	"host":          ".Host",
	"show_details":  ".Config.ShowRequestDetails",
	"hide_details":  "(not .Config.ShowRequestDetails)",
	"l10n_disabled": ".Config.L10nDisabled",
	"l10n_enabled":  "(not .Config.L10nDisabled)",
}

// v3tov4Fields maps old dot-field paths to new dot-field paths.
//
// Deprecated: Temporary workaround, used by [convertV3toV4].
var v3tov4Fields = map[string]string{ //nolint:gochecknoglobals
	".Code":               ".StatusCode",
	".ShowRequestDetails": ".Config.ShowRequestDetails",
	".L10nDisabled":       ".Config.L10nDisabled",
}

// Deprecated: those regexes are only used by [convertV3toV4].
var (
	actionRe = regexp.MustCompile(`\{\{([\s\S]*?)}}`)
	identRe  = regexp.MustCompile(`\.?[a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*`)
)

// convertV3toV4 takes a template source string and replaces all occurrences of old function-style tokens with
// their new dot-field path equivalents.
//
// Deprecated: This function is temporary workaround to give more time for users to migrate their templates to the
// new format. It will be removed in the future, so please update your templates to use the new dot-field syntax
// as soon as possible.
func convertV3toV4(src string) string {
	return actionRe.ReplaceAllStringFunc(src, func(action string) string {
		inner := actionRe.FindStringSubmatch(action)[1]
		replaced := identRe.ReplaceAllStringFunc(inner, func(ident string) string {
			if strings.HasPrefix(ident, ".") {
				if v, ok := v3tov4Fields[ident]; ok {
					return v
				}

				return ident
			}

			if v, ok := v3tov4Tokens[ident]; ok {
				return v
			}

			return ident
		})

		return "{{" + replaced + "}}"
	})
}
