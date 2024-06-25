package error_page

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/logger"
	"gh.tarampamp.am/error-pages/internal/template"
)

const contentTypeHeader = "Content-Type"

// New creates a new handler that returns an error page with the specified status code and format.
func New(cfg *config.Config, log *logger.Logger) http.Handler { //nolint:funlen,gocognit
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var code uint16

		if fromUrl, okUrl := extractCodeFromURL(r.URL.Path); okUrl {
			code = fromUrl
		} else if fromHeader, okHeaders := extractCodeFromHeaders(r.Header); okHeaders {
			code = fromHeader
		} else {
			code = cfg.DefaultCodeToRender
		}

		var httpCode int

		if cfg.RespondWithSameHTTPCode {
			httpCode = int(code)
		} else {
			httpCode = http.StatusOK
		}

		var format = detectPreferredFormatForClient(r.Header)

		{ // deal with the headers
			switch format {
			case jsonFormat:
				w.Header().Set(contentTypeHeader, "application/json; charset=utf-8")
			case xmlFormat:
				w.Header().Set(contentTypeHeader, "application/xml; charset=utf-8")
			case htmlFormat:
				w.Header().Set(contentTypeHeader, "text/html; charset=utf-8")
			default:
				w.Header().Set(contentTypeHeader, "text/plain; charset=utf-8") // plainTextFormat as default
			}

			// https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag
			// disallow indexing of the error pages
			w.Header().Set("X-Robots-Tag", "noindex")

			if code >= 500 && code < 600 {
				// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
				// tell the client (search crawler) to retry the request after 120 seconds, it makes sense for the 5xx errors
				w.Header().Set("Retry-After", "120")
			}

			// proxy the headers from the incoming request to the error page response if they are defined in the config
			for _, proxyHeader := range cfg.ProxyHeaders {
				if value := r.Header.Get(proxyHeader); value != "" {
					w.Header().Set(proxyHeader, value)
				}
			}
		}

		w.WriteHeader(httpCode)

		// prepare the template properties for rendering
		var tplProps = template.Props{
			Code:               code,             // http status code
			ShowRequestDetails: cfg.ShowDetails,  // status message
			L10nDisabled:       cfg.L10n.Disable, // status description
		}

		//nolint:lll
		if cfg.ShowDetails { // https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/
			tplProps.OriginalURI = r.Header.Get("X-Original-URI")   // (ingress-nginx) URI that caused the error
			tplProps.Namespace = r.Header.Get("X-Namespace")        // (ingress-nginx) namespace where the backend Service is located
			tplProps.IngressName = r.Header.Get("X-Ingress-Name")   // (ingress-nginx) name of the Ingress where the backend is defined
			tplProps.ServiceName = r.Header.Get("X-Service-Name")   // (ingress-nginx) name of the Service backing the backend
			tplProps.ServicePort = r.Header.Get("X-Service-Port")   // (ingress-nginx) port number of the Service backing the backend
			tplProps.RequestID = r.Header.Get("X-Request-Id")       // (ingress-nginx) unique ID that identifies the request - same as for backend service
			tplProps.ForwardedFor = r.Header.Get("X-Forwarded-For") // the value of the `X-Forwarded-For` header
			tplProps.Host = r.Header.Get("Host")                    // the value of the `Host` header
		}

		// try to find the code message and description in the config and if not - use the standard status text or fallback
		if desc, found := cfg.Codes.Find(code); found {
			tplProps.Message = desc.Message
			tplProps.Description = desc.Description
		} else if stdlibStatusText := http.StatusText(int(code)); stdlibStatusText != "" {
			tplProps.Message = stdlibStatusText
		} else {
			tplProps.Message = "Unknown Status Code" // fallback
		}

		switch {
		case format == jsonFormat && cfg.Formats.JSON != "":
			if content, err := template.Render(cfg.Formats.JSON, tplProps); err != nil {
				j, _ := json.Marshal(fmt.Sprintf("Failed to render the JSON template: %s", err.Error()))
				write(w, log, j)
			} else {
				write(w, log, content)
			}

		case format == xmlFormat && cfg.Formats.XML != "":
			if content, err := template.Render(cfg.Formats.XML, tplProps); err != nil {
				write(w, log, fmt.Sprintf(
					"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<error>Failed to render the XML template: %s</error>", err.Error(),
				))
			} else {
				write(w, log, content)
			}

		case format == htmlFormat:
			if tpl, found := cfg.Templates.Get(cfg.TemplateName); found {
				if content, err := template.Render(tpl, tplProps); err != nil {
					write(w, log, fmt.Sprintf(
						"<!DOCTYPE html>\n<html><body>Failed to render the HTML template %s: %s</body></html>",
						cfg.TemplateName,
						err.Error(),
					))
				} else {
					write(w, log, content)
				}
			} else {
				write(w, log, fmt.Sprintf(
					"<!DOCTYPE html>\n<html><body>Template %s not found and cannot be used</body></html>", cfg.TemplateName,
				))
			}

		default: // plainTextFormat as default
			if cfg.Formats.PlainText != "" {
				if content, err := template.Render(cfg.Formats.PlainText, tplProps); err != nil {
					write(w, log, fmt.Sprintf("Failed to render the PlainText template: %s", err.Error()))
				} else {
					write(w, log, content)
				}
			} else {
				write(w, log, `The requested content format is not supported.
Please create an issue on the project's GitHub page to request support for it.

Supported formats: JSON, XML, HTML, Plain Text`)
			}
		}
	})
}

func write[T string | []byte](w http.ResponseWriter, log *logger.Logger, content T) {
	var data []byte

	if s, ok := any(content).(string); ok {
		data = []byte(s)
	} else {
		data = any(content).([]byte)
	}

	if _, err := w.Write(data); err != nil && log != nil {
		log.Error("failed to write the response body",
			logger.String("content", string(data)),
			logger.Error(err),
		)
	}
}
