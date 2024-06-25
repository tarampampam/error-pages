package config

import (
	"maps"
	"net/http"
	"slices"

	builtinTemplates "gh.tarampamp.am/error-pages/templates"
)

type Config struct {
	// Templates hold all templates, with the key being the template name and the value being the template content
	// in HTML format (Go templates are supported here).
	Templates templates

	// Formats contain alternative response formats (e.g., if a client requests a response in one of these formats,
	// we will render the response using the specified format instead of HTML; Go templates are supported).
	Formats struct {
		JSON      string
		XML       string
		PlainText string
	}

	// Codes hold descriptions for HTTP codes (e.g., 404: "Not Found / The server can not find the requested page").
	Codes Codes

	// TemplateName is the name of the template to use for rendering error pages. The template must be present in the
	// Templates map.
	TemplateName string

	// ProxyHeaders contains a list of HTTP headers that will be proxied from the incoming request to the
	// error page response.
	ProxyHeaders []string

	// L10n contains localization settings.
	L10n struct {
		// Disable the localization of error pages.
		Disable bool
	}

	// DefaultCodeToRender is the code for the default error page to be displayed. It is used when the requested
	// code is not defined in the incoming request (i.e., the code to render as the index page).
	DefaultCodeToRender uint16

	// RespondWithSameHTTPCode determines whether the response should have the same HTTP status code as the requested
	// error page.
	// In other words, if set to true and the requested error page has a code of 404, the HTTP response will also have
	// a status code of 404. If set to false, the HTTP response will have a status code of 200 regardless of the
	// requested error page's status code.
	RespondWithSameHTTPCode bool

	// RotationMode allows to set the rotation mode for templates to switch between them automatically on startup,
	// on each request, daily, hourly and so on.
	RotationMode RotationMode

	// ShowDetails determines whether to show additional details in the error response, extracted from the
	// incoming request (if supported by the template).
	ShowDetails bool
}

const defaultJSONFormat string = `{
  "error": true,
  "code": {{ code | json }},
  "message": {{ message | json }},
  "description": {{ description | json }}{{ if show_details }},
  "details": {
    "host": {{ host | json }},
    "original_uri": {{ original_uri | json }},
    "forwarded_for": {{ forwarded_for | json }},
    "namespace": {{ namespace | json }},
    "ingress_name": {{ ingress_name | json }},
    "service_name": {{ service_name | json }},
    "service_port": {{ service_port | json }},
    "request_id": {{ request_id | json }},
    "timestamp": {{ now.Unix }}
  }{{ end }}
}`

const defaultXMLFormat string = `<?xml version="1.0" encoding="utf-8"?>
<error>
  <code>{{ code }}</code>
  <message>{{ message }}</message>
  <description>{{ description }}</description>{{ if show_details }}
  <details>
    <host>{{ host }}</host>
    <originalURI>{{ original_uri }}</originalURI>
    <forwardedFor>{{ forwarded_for }}</forwardedFor>
    <namespace>{{ namespace }}</namespace>
    <ingressName>{{ ingress_name }}</ingressName>
    <serviceName>{{ service_name }}</serviceName>
    <servicePort>{{ service_port }}</servicePort>
    <requestID>{{ request_id }}</requestID>
    <timestamp>{{ now.Unix }}</timestamp>
  </details>{{ end }}
</error>`

const defaultPlainTextFormat string = `Error {{ code }}: {{ message }}{{ if description }}
{{ description }}{{ end }}{{ if show_details }}

Host: {{ host }}
Original URI: {{ original_uri }}
Forwarded For: {{ forwarded_for }}
Namespace: {{ namespace }}
Ingress Name: {{ ingress_name }}
Service Name: {{ service_name }}
Service Port: {{ service_port }}
Request ID: {{ request_id }}
Timestamp: {{ now.Unix }}{{ end }}`

//nolint:lll
var defaultCodes = Codes{ //nolint:gochecknoglobals
	"400": {"Bad Request", "The server did not understand the request"},
	"401": {"Unauthorized", "The requested page needs a username and a password"},
	"403": {"Forbidden", "Access is forbidden to the requested page"},
	"404": {"Not Found", "The server can not find the requested page"},
	"405": {"Method Not Allowed", "The method specified in the request is not allowed"},
	"407": {"Proxy Authentication Required", "You must authenticate with a proxy server before this request can be served"},
	"408": {"Request Timeout", "The request took longer than the server was prepared to wait"},
	"409": {"Conflict", "The request could not be completed because of a conflict"},
	"410": {"Gone", "The requested page is no longer available"},
	"411": {"Length Required", "The \"Content-Length\" is not defined. The server will not accept the request without it"},
	"412": {"Precondition Failed", "The pre condition given in the request evaluated to false by the server"},
	"413": {"Payload Too Large", "The server will not accept the request, because the request entity is too large"},
	"416": {"Requested Range Not Satisfiable", "The requested byte range is not available and is out of bounds"},
	"418": {"I'm a teapot", "Attempt to brew coffee with a teapot is not supported"},
	"429": {"Too Many Requests", "Too many requests in a given amount of time"},
	"500": {"Internal Server Error", "The server met an unexpected condition"},
	"502": {"Bad Gateway", "The server received an invalid response from the upstream server"},
	"503": {"Service Unavailable", "The server is temporarily overloading or down"},
	"504": {"Gateway Timeout", "The gateway has timed out"},
	"505": {"HTTP Version Not Supported", "The server does not support the \"http protocol\" version"},
}

var defaultProxyHeaders = []string{ //nolint:gochecknoglobals
	// "Traceparent",  // W3C Trace Context
	// "Tracestate",   // W3C Trace Context
	"X-Request-Id",    // unofficial HTTP header, used to trace individual HTTP requests
	"X-Trace-Id",      // same as above
	"X-Amzn-Trace-Id", // to track HTTP requests from clients to targets or other AWS services
}

// New creates a new configuration with default values.
func New() Config {
	var cfg = Config{
		Templates: make(templates),          // allocate memory for templates
		Codes:     maps.Clone(defaultCodes), // copy default codes
	}

	cfg.Formats.JSON = defaultJSONFormat
	cfg.Formats.XML = defaultXMLFormat
	cfg.Formats.PlainText = defaultPlainTextFormat

	// add built-in templates
	for name, content := range builtinTemplates.BuiltIn() {
		cfg.Templates[name] = content
	}

	// set first template as default
	for _, name := range cfg.Templates.Names() {
		cfg.TemplateName = name

		break
	}

	// set default HTTP headers to proxy
	cfg.ProxyHeaders = slices.Clone(defaultProxyHeaders)

	// set defaults
	cfg.DefaultCodeToRender = http.StatusNotFound

	return cfg
}
