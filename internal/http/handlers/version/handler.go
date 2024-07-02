package version

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
)

// New creates a handler that returns the version of the service in JSON format.
func New(ver string) fasthttp.RequestHandler {
	var body, _ = json.Marshal(struct { //nolint:errchkjson
		Version string `json:"version"`
	}{
		Version: strings.TrimSpace(ver),
	})

	var notAllowed = http.StatusText(http.StatusMethodNotAllowed) + "\n"

	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			ctx.SetContentType("application/json; charset=utf-8")
			ctx.SetStatusCode(http.StatusOK)
			_, _ = ctx.Write(body)

		case fasthttp.MethodHead:
			ctx.SetStatusCode(http.StatusOK)

		default:
			ctx.Error(notAllowed, http.StatusMethodNotAllowed)
		}
	}
}
