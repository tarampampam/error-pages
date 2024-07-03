package static

import (
	_ "embed"
	"net/http"

	"github.com/valyala/fasthttp"
)

//go:embed favicon.ico
var Favicon []byte

// New creates a new handler that returns the provided content for GET and HEAD requests.
func New(content []byte) fasthttp.RequestHandler {
	var notAllowed = http.StatusText(http.StatusMethodNotAllowed) + "\n"

	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			ctx.SetContentType(http.DetectContentType(content))
			ctx.SetStatusCode(http.StatusOK)
			_, _ = ctx.Write(content)

		case fasthttp.MethodHead:
			ctx.SetStatusCode(http.StatusOK)

		default:
			ctx.Error(notAllowed, http.StatusMethodNotAllowed)
		}
	}
}
