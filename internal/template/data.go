package tpl

// Data represents the data structure that holds all the information needed to render an error page template.
//
// DO NOT MODIFY EXISTING FIELDS OR THEIR TYPES, as they are used in the templates and may be referenced in
// the template files.
//
// Note: After adding new fields, make sure to update the test data in [template_test.go] and add tests that verify
// the new fields are correctly rendered in the templates.
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
	HomepageURL  string // homepage URL (optional, set via --homepage-url)
	Links        []Link // additional links to display on the error page (optional, set via --add-link)
	Config       Config // configuration values

	// TODO: add incoming request headers as a map[string]string field, so they can be used in the templates?
}

// Link represents a labeled hyperlink that can be displayed in error page templates.
//
// DO NOT MODIFY EXISTING FIELDS OR THEIR TYPES.
type Link struct {
	Label string // link text shown to the user
	URL   string // target URL
}

// Config holds configuration values that can be used in the templates.
//
// DO NOT MODIFY EXISTING FIELDS OR THEIR TYPES.
type Config struct {
	ShowRequestDetails bool // show request details?
	L10nDisabled       bool // disable localization feature?
}
