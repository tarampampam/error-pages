package shared

import (
	"fmt"
	"strings"

	"gh.tarampamp.am/error-pages/v4/internal/cli"
	"gh.tarampamp.am/error-pages/v4/internal/codes"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
)

// NewDisableBuiltInCodesFlag returns a flag that disables the built-in HTTP status code descriptions.
func NewDisableBuiltInCodesFlag() cli.Flag[bool] {
	return cli.Flag[bool]{
		Names:   []string{"disable-built-in-codes"},
		Usage:   "Disable the built-in descriptions for HTTP status codes",
		EnvVars: []string{"DISABLE_BUILT_IN_CODES"},
		Default: false,
	}
}

// NewAddHTTPCodesFlag returns a flag for adding or overriding HTTP status codes and their descriptions.
func NewAddHTTPCodesFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names: []string{"add-code"},
		Usage: "Add or override HTTP status codes and their messages/descriptions " +
			"(format: 'CODE=MESSAGE[|DESCRIPTION][||CODE=MESSAGE[|DESCRIPTION]...]'; CODE may contain wildcards like " +
			"'4**'; separate multiple entries with '||', a newline, or a tab)",
		EnvVars: []string{"ADD_CODE"},
		Validator: func(_ *cli.Command, s string) error {
			_, err := ParseAddHTTPCodes(s)

			return err
		},
	}
}

// ParseAddHTTPCodes parses the --add-code flag value into a map of HTTP codes to their descriptions.
// Entries are separated by '||', newline, or tab; each entry has the format 'CODE=MESSAGE' or
// 'CODE=MESSAGE|DESCRIPTION'. Returns an error if any entry is malformed.
// Should be used together with [newAddHTTPCodesFlag].
func ParseAddHTTPCodes(s string) (map[string]codes.Description, error) {
	s = strings.ReplaceAll(s, "\n", "||")
	s = strings.ReplaceAll(s, "\t", "||")

	parts := strings.Split(s, "||")
	result := make(map[string]codes.Description, len(parts))

	for _, entry := range parts {
		if entry = strings.TrimSpace(entry); entry == "" {
			continue
		}

		before, after, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("wrong HTTP code entry %q: missing '='", entry)
		}

		code := strings.TrimSpace(before)
		if code == "" {
			return nil, fmt.Errorf("missing HTTP code in entry %q", entry)
		}

		if len(code) != 3 { //nolint:mnd
			return nil, fmt.Errorf("wrong HTTP code %q: must be 3 characters long", code)
		}

		for i := range len(code) {
			if b := code[i]; (b < '0' || b > '9') && b != '*' && b != 'x' && b != 'X' {
				return nil, fmt.Errorf("wrong HTTP code %q: allowed characters are digits and wildcards (*xX)", code)
			}
		}

		rest := after

		var msg, full string

		if before, after, ok = strings.Cut(rest, "|"); ok {
			msg, full = strings.TrimSpace(before), strings.TrimSpace(after)
		} else {
			msg = strings.TrimSpace(rest)
		}

		if msg == "" {
			return nil, fmt.Errorf("missing message for HTTP code %q", code)
		}

		result[code] = codes.Description{Short: msg, Full: full}
	}

	return result, nil
}

// NewHomepageURLFlag returns a flag for setting the homepage URL shown as a link in error pages.
func NewHomepageURLFlag(def string) cli.Flag[string] {
	return cli.Flag[string]{
		Names:   []string{"homepage-url"},
		Usage:   "Homepage URL to show as a link in error pages (e.g. https://app.example.com/home)",
		EnvVars: []string{"HOMEPAGE_URL"},
		Default: def,
	}
}

// NewAddLinksFlag returns a flag for adding extra labeled links to error pages.
func NewAddLinksFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names: []string{"add-link"},
		Usage: "Add extra links to error pages " +
			"(format: 'LABEL=URL[||LABEL=URL...]'; separate multiple entries with '||', a newline, or a tab)",
		EnvVars: []string{"ADD_LINK"},
		Validator: func(_ *cli.Command, s string) error {
			_, err := ParseLinks(s)

			return err
		},
	}
}

// ParseLinks parses the --add-link flag value into a slice of Link pairs.
// Entries are separated by '||', newline, or tab; each entry has the format 'LABEL=URL' where only
// the first '=' is used as the split point so that URLs containing '=' are handled correctly.
// Returns an error if any entry is malformed.
func ParseLinks(s string) ([]tpl.Link, error) {
	s = strings.ReplaceAll(s, "\n", "||")
	s = strings.ReplaceAll(s, "\t", "||")

	parts := strings.Split(s, "||")
	result := make([]tpl.Link, 0, len(parts))

	for _, entry := range parts {
		if entry = strings.TrimSpace(entry); entry == "" {
			continue
		}

		label, url, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("wrong link entry %q: missing '='", entry)
		}

		label = strings.TrimSpace(label)
		if label == "" {
			return nil, fmt.Errorf("missing label in link entry %q", entry)
		}

		url = strings.TrimSpace(url)
		if url == "" {
			return nil, fmt.Errorf("missing URL in link entry %q", entry)
		}

		result = append(result, tpl.Link{Label: label, URL: url})
	}

	return result, nil
}

// NewDisableL10nFlag returns a flag that disables client-side localization for templates that support it.
func NewDisableL10nFlag() cli.Flag[bool] {
	return cli.Flag[bool]{
		Names:   []string{"disable-l10n"},
		Usage:   "Disable localization of error pages (if the template supports localization)",
		EnvVars: []string{"DISABLE_L10N"},
		Default: false,
	}
}
