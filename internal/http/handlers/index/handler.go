package index

import (
	"github.com/tarampampam/error-pages/internal/config"
	"github.com/tarampampam/error-pages/internal/http/utils"
	"github.com/valyala/fasthttp"
)

type (
	templatePicker interface {
		// Pick the template name for responding.
		Pick() string
	}
)

// NewHandler creates handler for the index page serving.
func NewHandler(
	cfg *config.Config,
	p templatePicker,
	defaultPageCode string,
	defaultHTTPCode uint16,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		utils.RespondWithErrorPage(ctx, cfg, p, defaultPageCode, int(defaultHTTPCode))
	}
}
