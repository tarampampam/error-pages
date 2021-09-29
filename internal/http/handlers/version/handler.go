package version

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

// NewHandler creates version handler.
func NewHandler(ver string) fasthttp.RequestHandler {
	var cache []byte

	return func(ctx *fasthttp.RequestCtx) {
		if cache == nil {
			cache, _ = json.Marshal(struct {
				Version string `json:"version"`
			}{
				Version: ver,
			})
		}

		ctx.SetContentType("application/json")
		ctx.SetStatusCode(fasthttp.StatusOK)
		_, _ = ctx.Write(cache)
	}
}
