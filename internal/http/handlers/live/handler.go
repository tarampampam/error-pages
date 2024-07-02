package live

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

// New creates a new handler that returns "OK" for GET and HEAD requests.
func New() fasthttp.RequestHandler {
	var (
		body       = []byte("OK\n")
		notAllowed = http.StatusText(http.StatusMethodNotAllowed) + "\n"
	)

	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			ctx.SetContentType("text/plain; charset=utf-8")
			ctx.SetStatusCode(http.StatusOK)
			_, _ = ctx.Write(body)

		case fasthttp.MethodHead:
			ctx.SetStatusCode(http.StatusOK)

		default:
			ctx.Error(notAllowed, http.StatusMethodNotAllowed)
		}
	}
}
