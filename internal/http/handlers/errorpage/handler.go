package errorpage

import (
	"github.com/valyala/fasthttp"

	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/http/core"
	"gh.tarampamp.am/error-pages/internal/options"
	"gh.tarampamp.am/error-pages/internal/tpl"
)

type (
	templatePicker interface {
		// Pick the template name for responding.
		Pick() string
	}

	renderer interface {
		Render(content []byte, props tpl.Properties) ([]byte, error)
	}
)

// NewHandler creates handler for error pages serving.
func NewHandler(cfg *config.Config, p templatePicker, rdr renderer, opt options.ErrorPage) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		core.SetClientFormat(ctx, core.PlainTextContentType) // default content type

		if code, ok := ctx.UserValue("code").(string); ok {
			core.RespondWithErrorPage(ctx, cfg, p, rdr, code, fasthttp.StatusOK, opt)
		} else { // will never occur
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			_, _ = ctx.WriteString("cannot extract requested code from the request")
		}
	}
}
