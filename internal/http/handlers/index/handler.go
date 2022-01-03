package index

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

// NewHandler creates handler for the index page serving.
func NewHandler(
	e errorsPager,
	p templatePicker,
	defaultPageCode string,
	defaultHTTPCode uint16,
) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		content, err := e.GetPage(p.Pick(), defaultPageCode)

		if err == nil {
			ctx.SetContentType("text/html; charset=utf-8")
			ctx.SetStatusCode(int(defaultHTTPCode))
			_, _ = ctx.Write(content)

			return
		}

		ctx.SetContentType("text/plain; charset=utf-8")
		ctx.SetStatusCode(fasthttp.StatusNotAcceptable)
		_, _ = ctx.WriteString("default page code " + defaultPageCode + " is not available: " + err.Error())
	}
}
