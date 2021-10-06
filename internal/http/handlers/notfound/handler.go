package notfound

import (
	"github.com/valyala/fasthttp"
)

// NewHandler creates handler missing requests handling.
func NewHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("text/plain; charset=utf-8")
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		_, _ = ctx.WriteString("Wrong request URL. Error pages are available at the following URLs: /{code}.html")
	}
}
