package errorpage

import (
	"github.com/valyala/fasthttp"
)

type (
	errorsPager interface {
		// GetPage with passed template name and error code.
		GetPage(templateName, code string) ([]byte, error)
	}

	templatePicker interface {
		// Pick the template name for responding.
		Pick() string
	}
)

// NewHandler creates handler for error pages serving.
func NewHandler(e errorsPager, p templatePicker) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("text/plain; charset=utf-8") // default content type

		if code, ok := ctx.UserValue("code").(string); ok {
			if content, err := e.GetPage(p.Pick(), code); err == nil {
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetContentType("text/html; charset=utf-8")
				_, _ = ctx.Write(content)
			} else {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				_, _ = ctx.WriteString("requested code not available: " + err.Error()) // TODO customize the output?
			}
		} else { // will never happen
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			_, _ = ctx.WriteString("cannot extract requested code from the request") // TODO customize the output?
		}
	}
}
