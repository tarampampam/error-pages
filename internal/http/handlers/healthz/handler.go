package healthz

import "github.com/valyala/fasthttp"

// checker allows to check some service part.
type checker interface {
	// Check makes a check and return error only if something is wrong.
	Check() error
}

// NewHandler creates healthcheck handler.
func NewHandler(checker checker) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if err := checker.Check(); err != nil {
			ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
			_, _ = ctx.WriteString(err.Error())

			return
		}

		ctx.SetStatusCode(fasthttp.StatusOK)
		_, _ = ctx.WriteString("OK")
	}
}
