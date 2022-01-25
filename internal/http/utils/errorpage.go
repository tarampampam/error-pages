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

func RespondWithErrorPage(ctx *fasthttp.RequestCtx, cfg *config.Config, p templatePicker, code string, httpCode int) {
	var (
		_, canJSON = cfg.Formats[config.FormatJSON]
		props      = tpl.Properties{Code: code}
		clientWant = ClientWantFormat(ctx)
	)

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
			SetClientFormat(ctx, clientWant)

			if content, err := tpl.RenderJSON(cfg.Formats[config.FormatJSON].Content(), props); err == nil {
				ctx.SetStatusCode(httpCode)
				_, _ = ctx.Write(content)
			} else {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				_, _ = ctx.WriteString("cannot render JSON template: " + err.Error())
			}
		}

	default: // HTML
		{
			SetClientFormat(ctx, HTMLContentType)

			var templateName = p.Pick()

			if template, exists := cfg.Template(templateName); exists {
				if content, err := tpl.RenderHTML(template.Content(), props); err == nil {
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
