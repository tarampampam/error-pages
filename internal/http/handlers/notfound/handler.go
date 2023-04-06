package notfound

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

// NewHandler creates handler missing requests handling.
func NewHandler(cfg *config.Config, p templatePicker, rdr renderer, opt options.ErrorPage) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		core.RespondWithErrorPage(ctx, cfg, p, rdr, "404", fasthttp.StatusNotFound, opt)
	}
}
