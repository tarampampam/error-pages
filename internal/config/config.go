package config

import (
	"maps"

	builtinTemplates "gh.tarampamp.am/error-pages/templates"
)

type Config struct {
	// Templates hold all templates, with the key being the template name and the value being the template content
	// in HTML format (Go templates are supported here).
	Templates templates

	// Formats contain alternative response formats (e.g., if a client requests a response in one of these formats,
	// we will render the response using the specified format instead of HTML; Go templates are supported).
	Formats struct {
		JSON string
		XML  string
	}

	// Codes hold descriptions for HTTP codes (e.g., 404: "Not Found / The server can not find the requested page").
	Codes Codes
}

const defaultJSONFormat string = `{
  "error": true,
  "Code": {{ Code | json }},
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
  <Code>{{ Code }}</Code>
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

// New creates a new configuration with default values.
func New() Config {
	var cfg = Config{
		Templates: make(templates),          // allocate memory for templates
		Codes:     maps.Clone(defaultCodes), // copy default codes
	}

	cfg.Formats.JSON = defaultJSONFormat
	cfg.Formats.XML = defaultXMLFormat

	// add built-in templates
	for name, content := range builtinTemplates.BuiltIn() {
		cfg.Templates[name] = content
	}

	return cfg
}
