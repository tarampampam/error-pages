package shared

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/config"
)

const (
	CategoryHTTP      = "HTTP:"
	CategoryTemplates = "TEMPLATES:"
	CategoryCodes     = "HTTP CODES:"
	CategoryFormats   = "FORMATS:"
	CategoryBuild     = "BUILD:"
	CategoryOther     = "OTHER:"
)

// Note: Don't use pointers for flags, because they have own state which is not thread-safe.
// https://github.com/urfave/cli/issues/1926

var ListenAddrFlag = cli.StringFlag{
	Name:     "listen",
	Aliases:  []string{"l"},
	Usage:    "IP (v4 or v6) address to listen on",
	Value:    "0.0.0.0", // bind to all interfaces by default
	Sources:  cli.EnvVars("LISTEN_ADDR"),
	Category: CategoryHTTP,
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
	Validator: func(ip string) error {
		if ip == "" {
			return fmt.Errorf("missing IP address")
		}

		if net.ParseIP(ip) == nil {
			return fmt.Errorf("wrong IP address [%s] for listening", ip)
		}

		return nil
	},
}

var ListenPortFlag = cli.UintFlag{
	Name:     "port",
	Aliases:  []string{"p"},
	Usage:    "TCP port number",
	Value:    8080, // default port number
	Sources:  cli.EnvVars("LISTEN_PORT"),
	Category: CategoryHTTP,
	OnlyOnce: true,
	Validator: func(port uint) error {
		if port == 0 || port > 65535 {
			return fmt.Errorf("wrong TCP port number [%d]", port)
		}

		return nil
	},
}

var AddTemplatesFlag = cli.StringSliceFlag{
	Name: "add-template",
	Usage: "To add a new template, provide the path to the file using this flag (the filename without the extension " +
		"will be used as the template name)",
	Config:   cli.StringConfig{TrimSpace: true},
	Sources:  cli.EnvVars("ADD_TEMPLATE"),
	Category: CategoryTemplates,
	Validator: func(paths []string) error {
		for _, path := range paths {
			if path == "" {
				return fmt.Errorf("missing template path")
			}

			if stat, err := os.Stat(path); err != nil || stat.IsDir() {
				return fmt.Errorf("wrong template path [%s]", path)
			}
		}

		return nil
	},
}

var DisableTemplateNamesFlag = cli.StringSliceFlag{
	Name:     "disable-template",
	Usage:    "Disable the specified template by its name (useful to disable the built-in templates and use only custom ones)",
	Config:   cli.StringConfig{TrimSpace: true},
	Category: CategoryTemplates,
}

var AddHTTPCodesFlag = cli.StringMapFlag{
	Name: "add-code",
	Usage: "To add a new HTTP status code, provide the code and its message/description using this flag (the format " +
		"should be '%code%=%message%/%description%'; the code may contain a wildcard '*' to cover multiple codes at " +
		"once, for example, '4**' will cover all 4xx codes unless a more specific code is described previously)",
	Config:   cli.StringConfig{TrimSpace: true},
	Category: CategoryCodes,
	Validator: func(codes map[string]string) error {
		for code, msgAndDesc := range codes {
			if code == "" {
				return fmt.Errorf("missing HTTP code")
			} else if len(code) != 3 {
				return fmt.Errorf("wrong HTTP code [%s]: it should be 3 characters long", code)
			}

			if parts := strings.SplitN(msgAndDesc, "/", 3); len(parts) < 1 || len(parts) > 2 {
				return fmt.Errorf("wrong message/description format for HTTP code [%s]: %s", code, msgAndDesc)
			} else if parts[0] == "" {
				return fmt.Errorf("missing message for HTTP code [%s]", code)
			}
		}

		return nil
	},
}

// ParseHTTPCodes converts a map of HTTP status codes and their messages/descriptions into a map of codes and
// descriptions. Should be used together with [AddHTTPCodesFlag].
func ParseHTTPCodes(codes map[string]string) map[string]config.CodeDescription {
	var result = make(map[string]config.CodeDescription, len(codes))

	for code, msgAndDesc := range codes {
		var (
			parts = strings.SplitN(msgAndDesc, "/", 2)
			desc  config.CodeDescription
		)

		desc.Message = strings.TrimSpace(parts[0])

		if len(parts) > 1 {
			desc.Description = strings.TrimSpace(parts[1])
		}

		result[code] = desc
	}

	return result
}

var DisableL10nFlag = cli.BoolFlag{
	Name:     "disable-l10n",
	Usage:    "Disable localization of error pages (if the template supports localization)",
	Sources:  cli.EnvVars("DISABLE_L10N"),
	Category: CategoryOther,
	OnlyOnce: true,
}

var DisableMinificationFlag = cli.BoolFlag{
	Name:     "disable-minification",
	Usage:    "Disable the minification of HTML pages, including CSS, SVG, and JS (may be useful for debugging)",
	Sources:  cli.EnvVars("DISABLE_MINIFICATION"),
	Category: CategoryOther,
	OnlyOnce: true,
}
