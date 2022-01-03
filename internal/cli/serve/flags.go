package serve

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/tarampampam/error-pages/internal/env"
)

type flags struct {
	listen struct {
		ip   string
		port uint16
	}
	template struct {
		name string
	}
	defaultErrorPage string
	defaultHTTPCode  uint16
}

const (
	listenFlagName           = "listen"
	portFlagName             = "port"
	templateNameFlagName     = "template-name"
	defaultErrorPageFlagName = "default-error-page"
	defaultHTTPCodeFlagName  = "default-http-code"
)

const (
	useRandomTemplate              = "random"
	useRandomTemplateOnEachRequest = "i-said-random"
)

func (f *flags) init(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(
		&f.listen.ip,
		listenFlagName, "l",
		"0.0.0.0",
		fmt.Sprintf("IP address to listen on [$%s]", env.ListenAddr),
	)
	flagSet.Uint16VarP(
		&f.listen.port,
		portFlagName, "p",
		8080, //nolint:gomnd // must be same as default healthcheck `--port` flag value
		fmt.Sprintf("TCP port number [$%s]", env.ListenPort),
	)
	flagSet.StringVarP(
		&f.template.name,
		templateNameFlagName, "t",
		"",
		fmt.Sprintf(
			"template name (set \"%s\" to use a randomized or \"%s\" to use a randomized template on each request) [$%s]", //nolint:lll
			useRandomTemplate, useRandomTemplateOnEachRequest, env.TemplateName,
		),
	)
	flagSet.StringVarP(
		&f.defaultErrorPage,
		defaultErrorPageFlagName, "",
		"404",
		fmt.Sprintf("default error page [$%s]", env.DefaultErrorPage),
	)
	flagSet.Uint16VarP(
		&f.defaultHTTPCode,
		defaultHTTPCodeFlagName, "",
		404, //nolint:gomnd
		fmt.Sprintf("default HTTP response code [$%s]", env.DefaultHTTPCode),
	)
}

func (f *flags) overrideUsingEnv(flagSet *pflag.FlagSet) (lastErr error) { //nolint:gocognit
	flagSet.VisitAll(func(flag *pflag.Flag) {
		// flag was NOT defined using CLI (flags should have maximal priority)
		if !flag.Changed { //nolint:nestif
			switch flag.Name {
			case listenFlagName:
				if envVar, exists := env.ListenAddr.Lookup(); exists {
					f.listen.ip = strings.TrimSpace(envVar)
				}

			case portFlagName:
				if envVar, exists := env.ListenPort.Lookup(); exists {
					if p, err := strconv.ParseUint(envVar, 10, 16); err == nil { //nolint:gomnd
						f.listen.port = uint16(p)
					} else {
						lastErr = fmt.Errorf("wrong TCP port environment variable [%s] value", envVar)
					}
				}

			case templateNameFlagName:
				if envVar, exists := env.TemplateName.Lookup(); exists {
					f.template.name = strings.TrimSpace(envVar)
				}

			case defaultErrorPageFlagName:
				if envVar, exists := env.DefaultErrorPage.Lookup(); exists {
					f.defaultErrorPage = strings.TrimSpace(envVar)
				}

			case defaultHTTPCodeFlagName:
				if envVar, exists := env.DefaultHTTPCode.Lookup(); exists {
					if code, err := strconv.ParseUint(envVar, 10, 16); err == nil { //nolint:gomnd
						f.defaultHTTPCode = uint16(code)
					} else {
						lastErr = fmt.Errorf("wrong default HTTP response code environment variable [%s] value", envVar)
					}
				}
			}
		}
	})

	return lastErr
}

func (f *flags) validate() error {
	if net.ParseIP(f.listen.ip) == nil {
		return fmt.Errorf("wrong IP address [%s] for listening", f.listen.ip)
	}

	if f.defaultHTTPCode > 599 { //nolint:gomnd
		return fmt.Errorf("wrong default HTTP response code [%d]", f.defaultHTTPCode)
	}

	return nil
}
