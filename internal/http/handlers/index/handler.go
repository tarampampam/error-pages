package index

import (
	"strconv"

	"github.com/tarampampam/error-pages/internal/config"
	"github.com/tarampampam/error-pages/internal/http/core"
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
	showRequestDetails bool,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		pageCode, httpCode := defaultPageCode, int(defaultHTTPCode)

		if returnCode, ok := extractCodeToReturn(ctx); ok {
			pageCode, httpCode = strconv.Itoa(returnCode), returnCode
		}

		core.RespondWithErrorPage(ctx, cfg, p, pageCode, httpCode, showRequestDetails)
	}
}

func extractCodeToReturn(ctx *fasthttp.RequestCtx) (int, bool) { // for the Ingress support
	var ch = ctx.Request.Header.Peek(core.CodeHeader)

	if len(ch) > 0 && len(ch) <= 3 {
		if code, err := strconv.Atoi(string(ch)); err == nil {
			if code > 0 && code <= 599 {
				return code, true
			}
		}
	}

	return 0, false
}
