package core

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
	pageCode string,
	httpCode int,
	showRequestDetails bool,
) {
	ctx.Response.Header.Set("X-Robots-Tag", "noindex") // block Search indexing

	var (
		clientWant    = ClientWantFormat(ctx)
		json, canJSON = cfg.JSONFormat()
		xml, canXML   = cfg.XMLFormat()
		props         = tpl.Properties{Code: pageCode, ShowRequestDetails: showRequestDetails}
	)

	if showRequestDetails {
		props.OriginalURI = string(ctx.Request.Header.Peek(OriginalURI))
		props.Namespace = string(ctx.Request.Header.Peek(Namespace))
		props.IngressName = string(ctx.Request.Header.Peek(IngressName))
		props.ServiceName = string(ctx.Request.Header.Peek(ServiceName))
		props.ServicePort = string(ctx.Request.Header.Peek(ServicePort))
		props.RequestID = string(ctx.Request.Header.Peek(RequestID))
		props.ForwardedFor = string(ctx.Request.Header.Peek(ForwardedFor))
		props.Host = string(ctx.Request.Header.Peek(Host))
	}

	if page, exists := cfg.Pages[pageCode]; exists {
		props.Message = page.Message()
		props.Description = page.Description()
	} else if c, err := strconv.Atoi(pageCode); err == nil {
		if s := fasthttp.StatusMessage(c); s != "Unknown Status Code" { // as a fallback
			props.Message = s
		}
	}

	SetClientFormat(ctx, PlainTextContentType) // set default content type

	if props.Message == "" {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		_, _ = ctx.WriteString("requested pageCode (" + pageCode + ") not available")

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
