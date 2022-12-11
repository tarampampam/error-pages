package options

type ErrorPage struct {
	Default struct {
		PageCode string // default error page code
		HTTPCode uint16 // default HTTP response code
	}
	L10n struct {
		Disabled bool // disable error pages localization
	}
	Template struct {
		Name string // template name
	}
	ShowDetails      bool     // show request details in response
	CatchAll         bool     // catch every page with default http code and selected error page template
	ProxyHTTPHeaders []string // proxy HTTP request headers list
}
