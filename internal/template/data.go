package tpl

// Data represents the data structure that holds all the information needed to render an error page template.
//
// DO NOT MODIFY EXISTING FIELDS OR THEIR TYPES, as they are used in the templates and may be referenced in
// the template files.
type Data struct {
	StatusCode   uint16 // http status code
	Message      string // status message
	Description  string // status description
	OriginalURI  string // (ingress-nginx) URI that caused the error
	Namespace    string // (ingress-nginx) namespace where the backend Service is located
	IngressName  string // (ingress-nginx) name of the Ingress where the backend is defined
	ServiceName  string // (ingress-nginx) name of the Service backing the backend
	ServicePort  string // (ingress-nginx) port number of the Service backing the backend
	RequestID    string // (ingress-nginx, Envoy Gateway) unique ID that identifies the request
	ForwardedFor string // (ingress-nginx, Envoy Gateway) the value of the `X-Forwarded-For` header
	Host         string // the value of the `Host` header
	Config       Config // configuration values

	// TODO: add incoming request headers as a map[string]string field, so they can be used in the templates?
}

// Config holds configuration values that can be used in the templates.
//
// DO NOT MODIFY EXISTING FIELDS OR THEIR TYPES, as they are used in the templates and may be referenced in
// the template files.
type Config struct {
	ShowRequestDetails bool // show request details?
	L10nDisabled       bool // disable localization feature?
}
