package errorpage

import (
	"github.com/tarampampam/error-pages/internal/config"
	"github.com/tarampampam/error-pages/internal/http/core"
	"github.com/valyala/fasthttp"
)

type templatePicker interface {
	// Pick the template name for responding.
	Pick() string
}

// NewHandler creates handler for error pages serving.
func NewHandler(cfg *config.Config, p templatePicker, showRequestDetails bool) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		core.SetClientFormat(ctx, core.PlainTextContentType) // default content type

		if code, ok := ctx.UserValue("code").(string); ok {
			core.RespondWithErrorPage(ctx, cfg, p, code, fasthttp.StatusOK, showRequestDetails)
		} else { // will never occur
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			_, _ = ctx.WriteString("cannot extract requested code from the request")
		}
	}
}
