package template

import "reflect"

//nolint:lll
type Props struct {
	Code               uint16 `token:"code"`          // http status code
	Message            string `token:"message"`       // status message
	Description        string `token:"description"`   // status description
	OriginalURI        string `token:"original_uri"`  // (ingress-nginx) URI that caused the error
	Namespace          string `token:"namespace"`     // (ingress-nginx) namespace where the backend Service is located
	IngressName        string `token:"ingress_name"`  // (ingress-nginx) name of the Ingress where the backend is defined
	ServiceName        string `token:"service_name"`  // (ingress-nginx) name of the Service backing the backend
	ServicePort        string `token:"service_port"`  // (ingress-nginx) port number of the Service backing the backend
	RequestID          string `token:"request_id"`    // (ingress-nginx) unique ID that identifies the request - same as for backend service
	ForwardedFor       string `token:"forwarded_for"` // the value of the `X-Forwarded-For` header
	Host               string `token:"host"`          // the value of the `Host` header
	ShowRequestDetails bool   `token:"show_details"`  // (config) show request details?
	L10nDisabled       bool   `token:"l10n_disabled"` // (config) disable localization feature?
}

// Values convert the Props struct into a map where each key is a token associated with its corresponding value.
func (p Props) Values() map[string]any {
	var result = make(map[string]any, reflect.ValueOf(p).NumField())

	for i, v := 0, reflect.ValueOf(p); i < v.NumField(); i++ {
		if token, tagExists := v.Type().Field(i).Tag.Lookup("token"); tagExists {
			result[token] = v.Field(i).Interface()
		}
	}

	return result
}
