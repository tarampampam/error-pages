package utils

import (
	"strconv"

	"github.com/tarampampam/error-pages/internal/config"
	"github.com/tarampampam/error-pages/internal/tpl"
	"github.com/valyala/fasthttp"
)

type templatePicker interface {
	// Pick the template name for responding.
	Pick() string
}

func RespondWithErrorPage( //nolint:funlen
	ctx *fasthttp.RequestCtx,
	cfg *config.Config,
	p templatePicker,
	code string,
	httpCode int,
	showRequestDetails bool,
) {
	ctx.Response.Header.Set("X-Robots-Tag", "noindex") // block Search indexing

	if returnCode, ok := extractCodeToReturn(ctx); ok {
		code, httpCode = strconv.Itoa(returnCode), returnCode
	}

	var (
		clientWant    = ClientWantFormat(ctx)
		json, canJSON = cfg.JSONFormat()
		xml, canXML   = cfg.XMLFormat()
		props         = tpl.Properties{Code: code, ShowRequestDetails: showRequestDetails}
	)

	if showRequestDetails {
		props.OriginalURI = string(ctx.Request.Header.Peek(OriginalURI))
		props.Namespace = string(ctx.Request.Header.Peek(Namespace))
		props.IngressName = string(ctx.Request.Header.Peek(IngressName))
		props.ServiceName = string(ctx.Request.Header.Peek(ServiceName))
		props.ServicePort = string(ctx.Request.Header.Peek(ServicePort))
		props.RequestID = string(ctx.Request.Header.Peek(RequestID))
	}

	if page, exists := cfg.Pages[code]; exists {
		props.Message = page.Message()
		props.Description = page.Description()
	} else if c, err := strconv.Atoi(code); err == nil {
		if s := fasthttp.StatusMessage(c); s != "Unknown Status Code" { // as a fallback
			props.Message = s
		}
	}

	SetClientFormat(ctx, PlainTextContentType) // set default content type

	if props.Message == "" {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		_, _ = ctx.WriteString("requested code (" + code + ") not available")

		return
	}

	switch {
	case clientWant == JSONContentType && canJSON: // JSON
		{
			SetClientFormat(ctx, JSONContentType)

			if content, err := tpl.Render(json.Content(), props); err == nil {
				ctx.SetStatusCode(httpCode)
				_, _ = ctx.Write(content)
			} else {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				_, _ = ctx.WriteString("cannot render JSON template: " + err.Error())
			}
		}

	case clientWant == XMLContentType && canXML: // XML
		{
			SetClientFormat(ctx, XMLContentType)

			if content, err := tpl.Render(xml.Content(), props); err == nil {
				ctx.SetStatusCode(httpCode)
				_, _ = ctx.Write(content)
			} else {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				_, _ = ctx.WriteString("cannot render XML template: " + err.Error())
			}
		}

	default: // HTML
		{
			SetClientFormat(ctx, HTMLContentType)

			var templateName = p.Pick()

			if template, exists := cfg.Template(templateName); exists {
				if content, err := tpl.Render(template.Content(), props); err == nil {
					ctx.SetStatusCode(httpCode)
					_, _ = ctx.Write(content)
				} else {
					ctx.SetStatusCode(fasthttp.StatusInternalServerError)
					_, _ = ctx.WriteString("cannot render HTML template: " + err.Error())
				}
			} else {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				_, _ = ctx.WriteString("template " + templateName + " not exists")
			}
		}
	}
}

func extractCodeToReturn(ctx *fasthttp.RequestCtx) (int, bool) { // for the Ingress support
	var ch = ctx.Request.Header.Peek(CodeHeader)

	if len(ch) > 0 && len(ch) <= 3 {
		if code, err := strconv.Atoi(string(ch)); err == nil {
			if code > 0 && code <= 599 {
				return code, true
			}
		}
	}

	return 0, false
}
