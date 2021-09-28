package errorpage

import (
	"math/rand"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/tarampampam/error-pages/internal/http/common"
	"github.com/tarampampam/error-pages/internal/tpl"
	"github.com/valyala/fasthttp"
)

const (
	UseRandom              = "random"
	UseRandomOnEachRequest = "i-said-random"
)

// NewHandler creates handler for error pages serving.
func NewHandler(
	templateName string,
	templates map[string][]byte,
	codes map[string]tpl.Annotator,
) (fasthttp.RequestHandler, error) {
	if len(templates) == 0 {
		return nil, errors.New("empty templates map")
	}

	var (
		rnd           = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
		templateNames = templateTames(templates)
	)

	if templateName == "" { // on empty template name
		templateName = templateNames[0] // pick the first
	} else if templateName == UseRandom { // on "random" template name
		templateName = templateNames[rnd.Intn(len(templateNames))] // pick the randomized
	}

	if _, found := templates[templateName]; !found && templateName != UseRandomOnEachRequest {
		return nil, errors.New("wrong template name passed")
	}

	var pages = tpl.NewErrors(templates, codes)

	return func(ctx *fasthttp.RequestCtx) {
		var useTemplate = templateName // default

		if templateName == UseRandomOnEachRequest {
			useTemplate = templateNames[rnd.Intn(len(templateNames))] // pick the randomized
		}

		userCode := ctx.UserValue("code")

		if code, ok := userCode.(string); ok {
			if content, err := pages.Get(useTemplate, code); err == nil {
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetContentType("text/html; charset=utf-8")
				_, _ = ctx.Write(content)
			} else {
				common.HandleInternalHTTPError(
					ctx,
					fasthttp.StatusNotFound,
					"requested code not available: "+err.Error(),
				)
			}
		} else { // will never happen
			common.HandleInternalHTTPError(
				ctx,
				fasthttp.StatusInternalServerError,
				"cannot extract requested code from the request",
			)
		}
	}, nil
}

func templateTames(templates map[string][]byte) []string {
	var templateNames = make([]string, 0, len(templates))

	for name := range templates {
		templateNames = append(templateNames, name)
	}

	sort.Strings(templateNames)

	return templateNames
}
