package core

const (
	// FormatHeader name of the header used to extract the format
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP status code to return
	CodeHeader = "X-Code"

	// OriginalURI name of the header with the original URL from NGINX
	OriginalURI = "X-Forwarded-Host"

	// Namespace name of the header that contains information about the Ingress namespace
	Namespace = "X-Namespace"

	// IngressName name of the header that contains the matched Ingress
	IngressName = "X-Ingress-Name"

	// ServiceName name of the header that contains the matched Service in the Ingress
	ServiceName = "X-Forwarded-Server"

	// ServicePort name of the header that contains the matched Service port in the Ingress
	ServicePort = "X-Service-Port"

	//// RequestID is a unique ID that identifies the request - same as for backend service
	// RequestID = "X-Request-ID"

	// ForwardedFor identifies the user of this session
	ForwardedFor = "Cf-Connecting-Ip"

	// DataCenter Datacenter
	DataCenter = "Cf-Ipcity"

	// Proto is the protocol used by the client
	Proto = "X-Forwarded-Proto"

	// RayID is the unique ID of the request
	RayID = "Cf-Ray"

	//// Host identifies the hosts origin
	// Host = "Host"
)
