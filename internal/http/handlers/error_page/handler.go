package error_page

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"

	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/logger"
	"gh.tarampamp.am/error-pages/internal/template"
)

// New creates a new handler that returns an error page with the specified status code and format.
func New(cfg *config.Config, log *logger.Logger) (_ fasthttp.RequestHandler, closeCache func()) { //nolint:funlen,gocognit,gocyclo,lll
	// if the ttl will be bigger than 1 second, the template functions like `nowUnix` will not work as expected
	const cacheTtl = 900 * time.Millisecond // the cache TTL

	var (
		cache, stopCh = NewRenderedCache(cacheTtl), make(chan struct{})
		stopOnce      sync.Once
	)

	// run a goroutine that will clear the cache from expired items. to stop the goroutine - close the stop channel
	// or call the closeCache
	go func() {
		var timer = time.NewTimer(cacheTtl)

		defer func() { timer.Stop(); cache.Clear() }()

		for {
			select {
			case <-timer.C:
				cache.ClearExpired()
				timer.Reset(cacheTtl)
			case <-stopCh:
				return
			}
		}
	}()

	return func(ctx *fasthttp.RequestCtx) {
		var (
			reqHeaders = &ctx.Request.Header
			code       uint16
		)

		if fromUrl, okUrl := extractCodeFromURL(string(ctx.Path())); okUrl {
			code = fromUrl
		} else if fromHeader, okHeaders := extractCodeFromHeaders(reqHeaders); okHeaders {
			code = fromHeader
		} else {
			code = cfg.DefaultCodeToRender
		}

		var httpCode int

		if cfg.RespondWithSameHTTPCode {
			httpCode = int(code)
		} else {
			httpCode = http.StatusOK
		}

		var format = detectPreferredFormatForClient(reqHeaders)

		{ // deal with the headers
			switch format {
			case jsonFormat:
				ctx.SetContentType("application/json; charset=utf-8")
			case xmlFormat:
				ctx.SetContentType("application/xml; charset=utf-8")
			case htmlFormat:
				ctx.SetContentType("text/html; charset=utf-8")
			default:
				ctx.SetContentType("text/plain; charset=utf-8") // plainTextFormat as default
			}

			// https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag
			// disallow indexing of the error pages
			ctx.Response.Header.Set("X-Robots-Tag", "noindex")

			switch code {
			case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests,
				http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable,
				http.StatusGatewayTimeout:
				// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
				// tell the client (search crawler) to retry the request after 120 seconds
				ctx.Response.Header.Set("Retry-After", "120")
			}

			// proxy the headers from the incoming request to the error page response if they are defined in the config
			for _, proxyHeader := range cfg.ProxyHeaders {
				if value := reqHeaders.Peek(proxyHeader); len(value) > 0 {
					ctx.Response.Header.SetBytesV(proxyHeader, value)
				}
			}
		}

		ctx.SetStatusCode(httpCode)

		// prepare the template properties for rendering
		var tplProps = template.Props{
			Code:               code,             // http status code
			ShowRequestDetails: cfg.ShowDetails,  // status message
			L10nDisabled:       cfg.L10n.Disable, // status description
		}

		//nolint:lll
		if cfg.ShowDetails { // https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/
			tplProps.OriginalURI = string(reqHeaders.Peek("X-Original-URI"))   // (ingress-nginx) URI that caused the error
			tplProps.Namespace = string(reqHeaders.Peek("X-Namespace"))        // (ingress-nginx) namespace where the backend Service is located
			tplProps.IngressName = string(reqHeaders.Peek("X-Ingress-Name"))   // (ingress-nginx) name of the Ingress where the backend is defined
			tplProps.ServiceName = string(reqHeaders.Peek("X-Service-Name"))   // (ingress-nginx) name of the Service backing the backend
			tplProps.ServicePort = string(reqHeaders.Peek("X-Service-Port"))   // (ingress-nginx) port number of the Service backing the backend
			tplProps.RequestID = string(reqHeaders.Peek("X-Request-Id"))       // (ingress-nginx) unique ID that identifies the request - same as for backend service
			tplProps.ForwardedFor = string(reqHeaders.Peek("X-Forwarded-For")) // the value of the `X-Forwarded-For` header
			tplProps.Host = string(reqHeaders.Peek("Host"))                    // the value of the `Host` header
		}

		// try to find the code message and description in the config and if not - use the standard status text or fallback
		if desc, found := cfg.Codes.Find(code); found {
			tplProps.Message = desc.Message
			tplProps.Description = desc.Description
		} else if stdlibStatusText := http.StatusText(int(code)); stdlibStatusText != "" {
			tplProps.Message = stdlibStatusText
		} else {
			tplProps.Message = "Unknown Status Code" // fallback
		}

		switch {
		case format == jsonFormat && cfg.Formats.JSON != "":
			if cached, ok := cache.Get(cfg.Formats.JSON, tplProps); ok { // cache hit
				write(ctx, log, cached)
			} else { // cache miss
				if content, err := template.Render(cfg.Formats.JSON, tplProps); err != nil {
					errAsJson, _ := json.Marshal(fmt.Sprintf("Failed to render the JSON template: %s", err.Error()))
					write(ctx, log, errAsJson) // error during rendering
				} else {
					cache.Put(cfg.Formats.JSON, tplProps, []byte(content))

					write(ctx, log, content) // rendered successfully
				}
			}

		case format == xmlFormat && cfg.Formats.XML != "":
			if cached, ok := cache.Get(cfg.Formats.XML, tplProps); ok { // cache hit
				write(ctx, log, cached)
			} else { // cache miss
				if content, err := template.Render(cfg.Formats.XML, tplProps); err != nil {
					write(ctx, log, fmt.Sprintf(
						"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<error>Failed to render the XML template: %s</error>\n", err.Error(),
					))
				} else {
					cache.Put(cfg.Formats.XML, tplProps, []byte(content))

					write(ctx, log, content)
				}
			}

		case format == htmlFormat:
			var templateName = templateToUse(cfg)

			if tpl, found := cfg.Templates.Get(templateName); found { //nolint:nestif
				if cached, ok := cache.Get(tpl, tplProps); ok { // cache hit
					write(ctx, log, cached)
				} else { // cache miss
					if content, err := template.Render(tpl, tplProps); err != nil {
						// TODO: add GZIP compression for the HTML content support
						write(ctx, log, fmt.Sprintf(
							"<!DOCTYPE html>\n<html><body>Failed to render the HTML template %s: %s</body></html>\n",
							templateName,
							err.Error(),
						))
					} else {
						if !cfg.DisableMinification {
							if mini, minErr := template.MiniHTML(content); minErr != nil {
								log.Warn("HTML minification failed", logger.Error(minErr))
							} else {
								content = mini
							}
						}

						cache.Put(tpl, tplProps, []byte(content))

						write(ctx, log, content)
					}
				}
			} else {
				write(ctx, log, fmt.Sprintf(
					"<!DOCTYPE html>\n<html><body>Template %s not found and cannot be used</body></html>\n", templateName,
				))
			}

		default: // plainTextFormat as default
			if cfg.Formats.PlainText != "" { //nolint:nestif
				if cached, ok := cache.Get(cfg.Formats.PlainText, tplProps); ok { // cache hit
					write(ctx, log, cached)
				} else { // cache miss
					if content, err := template.Render(cfg.Formats.PlainText, tplProps); err != nil {
						write(ctx, log, fmt.Sprintf("Failed to render the PlainText template: %s", err.Error()))
					} else {
						cache.Put(cfg.Formats.PlainText, tplProps, []byte(content))

						write(ctx, log, content)
					}
				}
			} else {
				write(ctx, log, `The requested content format is not supported.
Please create an issue on the project's GitHub page to request support for this format.

Supported formats: JSON, XML, HTML, Plain Text
`)
			}
		}
	}, func() { stopOnce.Do(func() { close(stopCh) }) }
}

var (
	templateChangedAt atomic.Pointer[time.Time] //nolint:gochecknoglobals // the time when the theme was changed last time
	pickedTemplate    atomic.Pointer[string]    //nolint:gochecknoglobals // the name of the randomly picked template
)

// templateToUse decides which template to use based on the rotation mode and the last time the template was changed.
func templateToUse(cfg *config.Config) string {
	switch rotationMode := cfg.RotationMode; rotationMode {
	case config.RotationModeDisabled:
		return cfg.TemplateName // not needed to do anything
	case config.RotationModeRandomOnStartup:
		return cfg.TemplateName // do nothing, the scope of this rotation mode is not here
	case config.RotationModeRandomOnEachRequest:
		return cfg.Templates.RandomName() // pick a random template on each request
	case config.RotationModeRandomHourly, config.RotationModeRandomDaily:
		var now, rndTemplate = time.Now(), cfg.Templates.RandomName()

		if changedAt := templateChangedAt.Load(); changedAt == nil {
			// the template was not changed yet (first request)
			templateChangedAt.Store(&now)
			pickedTemplate.Store(&rndTemplate)

			return rndTemplate
		} else {
			// is it time to change the template?
			if (rotationMode == config.RotationModeRandomHourly && changedAt.Hour() != now.Hour()) ||
				(rotationMode == config.RotationModeRandomDaily && changedAt.Day() != now.Day()) {
				templateChangedAt.Store(&now)
				pickedTemplate.Store(&rndTemplate)

				return rndTemplate
			} else if lastUsed := pickedTemplate.Load(); lastUsed != nil {
				// time to change the template has not come yet, so use the last picked template
				return *lastUsed
			} else {
				// in case if the last picked template is not set, pick a random one and store it
				templateChangedAt.Store(&now)
				pickedTemplate.Store(&rndTemplate)

				return rndTemplate
			}
		}
	}

	return cfg.TemplateName // the fallback of the fallback :D
}

// write the content to the response writer and log the error if any.
func write[T string | []byte](ctx *fasthttp.RequestCtx, log *logger.Logger, content T) {
	var data []byte

	if s, ok := any(content).(string); ok {
		data = []byte(s)
	} else {
		data = any(content).([]byte)
	}

	if _, err := ctx.Write(data); err != nil && log != nil {
		log.Error("failed to write the response body",
			logger.String("content", string(data)),
			logger.Error(err),
		)
	}
}
