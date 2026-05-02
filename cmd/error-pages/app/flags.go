package app

import (
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
	"unicode"

	"gh.tarampamp.am/error-pages/v4/internal/cli"
	"gh.tarampamp.am/error-pages/v4/internal/logger"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/template/tploader"
)

func newLogLevelFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names: []string{"log-level"},
		Usage: "Logging level (" + strings.Join([]string{
			logger.DebugLevel.String(),
			logger.InfoLevel.String(),
			logger.WarnLevel.String(),
			logger.ErrorLevel.String(),
		}, "/") + ")",
		EnvVars: []string{"LOG_LEVEL"},
		Default: logger.InfoLevel.String(),
		Validator: func(_ *cli.Command, lvl string) (err error) {
			_, err = logger.ParseLevel(lvl)

			return
		},
	}
}

func newLogFormatFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names: []string{"log-format"},
		Usage: "Logging format (" + strings.Join([]string{
			logger.ConsoleFormat.String(),
			logger.JSONFormat.String(),
		}, "/") + ")",
		EnvVars: []string{"LOG_FORMAT"},
		Default: logger.ConsoleFormat.String(),
		Validator: func(_ *cli.Command, fmt string) (err error) {
			_, err = logger.ParseFormat(fmt)

			return
		},
	}
}

func newHTTPAddrFlag(def string) cli.Flag[string] {
	return cli.Flag[string]{
		Names:   []string{"addr", "listen"},
		Usage:   "HTTP server address to listen on (IPv4 or IPv6)",
		EnvVars: []string{"HTTP_ADDR", "LISTEN_ADDR", "ADDR"},
		Default: def,
		Validator: func(_ *cli.Command, ip string) error {
			if ip == "" {
				return errors.New("missing IP address for listening")
			}

			if net.ParseIP(ip) == nil {
				return fmt.Errorf("wrong IP address [%s] for listening", ip)
			}

			return nil
		},
	}
}

func newHTTPPortFlag(def uint) cli.Flag[uint] {
	return cli.Flag[uint]{
		Names:   []string{"port"},
		Usage:   "HTTP server TCP port number",
		EnvVars: []string{"HTTP_PORT", "LISTEN_PORT", "PORT"},
		Default: def,
		Validator: func(_ *cli.Command, port uint) error {
			if port == 0 || port > 65535 {
				return fmt.Errorf("wrong TCP port number [%d]", port)
			}

			return nil
		},
	}
}

func newDefaultCodeToRenderFlag(def uint) cli.Flag[uint] {
	return cli.Flag[uint]{
		Names:   []string{"default-error-page"},
		Usage:   "Default HTTP status code to render",
		EnvVars: []string{"DEFAULT_ERROR_PAGE"},
		Default: def,
		Validator: func(_ *cli.Command, code uint) error {
			if code > 999 { //nolint:mnd
				return fmt.Errorf("wrong HTTP code [%d] for the default error page", code)
			}

			return nil
		},
	}
}

func newSendSameHTTPCodeFlag() cli.Flag[bool] {
	return cli.Flag[bool]{
		Names:   []string{"send-same-http-code"},
		Usage:   "The HTTP response should use the same status code as the requested error page",
		EnvVars: []string{"SEND_SAME_HTTP_CODE"},
		Default: false,
	}
}

func newShowDetailsFlag() cli.Flag[bool] {
	return cli.Flag[bool]{
		Names:   []string{"show-details"},
		Usage:   "Show details about the request in the error page response (if supported by the template)",
		EnvVars: []string{"SHOW_DETAILS"},
		Default: false,
	}
}

func newProxyHeadersListFlag(def []string) cli.Flag[string] {
	return cli.Flag[string]{
		Names: []string{"proxy-headers"},
		Usage: "HTTP headers listed here will be proxied from the original request to the error page response " +
			"(comma/new-line separated list)",
		EnvVars: []string{"PROXY_HTTP_HEADERS"},
		Default: strings.Join(def, ","),
		Validator: func(_ *cli.Command, s string) error {
			for _, name := range splitProxyHeadersList(s) {
				for _, c := range name {
					// RFC 7230 #3.2.6: tchar = ALPHA / DIGIT /
					//   "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~"
					alphaNum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')

					special := strings.ContainsRune("!#$%&'*+-.^_`|~", c)
					if !alphaNum && !special {
						return fmt.Errorf("invalid HTTP header name %q", name)
					}
				}
			}

			return nil
		},
	}
}

// splitProxyHeadersList takes a comma/semicolon/space-separated list of HTTP header names, normalizes it by
// trimming whitespace and removing duplicates, and returns a slice of unique header names.
func splitProxyHeadersList(headers string) []string {
	if headers == "" {
		return nil
	}

	parts := strings.FieldsFunc(headers, func(r rune) bool { return r == ',' || r == ';' || unicode.IsSpace(r) })

	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if _, exists := seen[part]; !exists {
			seen[part] = struct{}{}
			result = append(result, part)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func newTemplateNameFlag(all []string, def string) cli.Flag[string] {
	return cli.Flag[string]{
		Names: []string{"template-name"},
		Usage: "Name of the built-in HTML template to use (" + strings.Join(all, "/") +
			"; ignored if a custom HTML template is set)",
		EnvVars: []string{"TEMPLATE_NAME", "HTML_TEMPLATE_NAME"},
		Default: def,
		Validator: func(_ *cli.Command, name string) error {
			if slices.Contains(all, name) {
				return nil
			}

			return fmt.Errorf("unknown built-in HTML template name %q (available templates: %s)", name, strings.Join(all, ", "))
		},
	}
}

func newRotationModeFlag(def tpl.RotationMode) cli.Flag[string] {
	all := []string{
		string(tpl.RotationModeDisabled),
		string(tpl.RotationModeRandomOnStartup),
		string(tpl.RotationModeRandomOnEachRequest),
		string(tpl.RotationModeRandomHourly),
		string(tpl.RotationModeRandomDaily),
	}

	return cli.Flag[string]{
		Names: []string{"rotation-mode"},
		Usage: "Mode for rotating built-in HTML templates (" + strings.Join(all, "/") +
			"; ignored if a custom HTML template is set)",
		EnvVars: []string{"ROTATION_MODE"},
		Default: string(def),
		Validator: func(_ *cli.Command, mode string) error {
			if slices.Contains(all, mode) {
				return nil
			}

			return fmt.Errorf("unknown rotation mode %q (available modes: %s)", mode, strings.Join(all, ", "))
		},
	}
}

func newHTMLTemplateFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names:     []string{"html-template"},
		Usage:     "Custom HTML template for error page responses (template text/URL/file path)",
		EnvVars:   []string{"HTML_TEMPLATE", "TEMPLATE"},
		Validator: validateCustomTemplate,
	}
}

func newJSONTemplateFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names:     []string{"json-template"},
		Usage:     "Custom JSON template for error page responses (template text/URL/file path)",
		EnvVars:   []string{"JSON_TEMPLATE"},
		Validator: validateCustomTemplate,
	}
}

func newXMLTemplateFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names:     []string{"xml-template"},
		Usage:     "Custom XML template for error page responses (template text/URL/file path)",
		EnvVars:   []string{"XML_TEMPLATE"},
		Validator: validateCustomTemplate,
	}
}

func newPlainTextTemplateFlag() cli.Flag[string] {
	return cli.Flag[string]{
		Names:     []string{"plaintext-template"},
		Usage:     "Custom plain text template for error page responses (template text/URL/file path)",
		EnvVars:   []string{"TEXT_TEMPLATE", "PLAINTEXT_TEMPLATE"},
		Validator: validateCustomTemplate,
	}
}

func validateCustomTemplate(_ *cli.Command, src string) error {
	if tploader.IsURL(src) || tploader.IsFilePath(src) {
		// if it's a URL or file path, we will attempt to load it later, so just skip validation for now
		return nil
	}

	t, err := tpl.New(src)
	if err != nil {
		return fmt.Errorf("custom template parsing: %w", err)
	}

	if err = t.RenderTo(tpl.Data{}, io.Discard); err != nil {
		return fmt.Errorf("custom template rendering test: %w", err)
	}

	return nil
}
