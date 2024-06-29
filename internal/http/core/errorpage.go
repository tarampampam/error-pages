package core

import (
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"

	"gh.tarampamp.am/error-pages/internal/config"
	"gh.tarampamp.am/error-pages/internal/options"
	"gh.tarampamp.am/error-pages/internal/tpl"
)

type templatePicker interface {
	// Pick the template name for responding.
	Pick() string
}

type renderer interface {
	Render(content []byte, props tpl.Properties) ([]byte, error)
}

// GenHostName 生成 hostName 的函数 随机生成类似country + c8824f96ccdb的字符串 country 是 header 中的 Cf-Ipcountry 如果没有则默认为 CN
func GenHostName(ctx *fasthttp.RequestCtx) string {
	var country = string(ctx.Request.Header.Peek(DataCenter))

	if country == "" {
		country = "CN"
	}

	// 增加处理 如果 country 带有空格 删掉空格
	return strings.ReplaceAll(country, " ", "") + "-" + uuid.New().String()[:12]

}

func RespondWithErrorPage( //nolint:funlen,gocyclo
	ctx *fasthttp.RequestCtx,
	cfg *config.Config,
	p templatePicker,
	rdr renderer,
	pageCode string,
	httpCode int,
	opt options.ErrorPage,
) {
	ctx.Response.Header.Set("X-Robots-Tag", "noindex") // block Search indexing

	var (
		clientWant    = ClientWantFormat(ctx)
		json, canJSON = cfg.JSONFormat()
		xml, canXML   = cfg.XMLFormat()
		props         = tpl.Properties{
			Code:               pageCode,
			ShowRequestDetails: opt.ShowDetails,
			L10nDisabled:       opt.L10n.Disabled,
		}
	)

	if opt.ShowDetails {
		props.OriginalURI = string(ctx.Request.Header.Peek(OriginalURI))
		props.Namespace = string(ctx.Request.Header.Peek(Namespace))
		props.IngressName = string(ctx.Request.Header.Peek(IngressName))
		props.ServiceName = string(ctx.Request.Header.Peek(ServiceName))
		props.ServicePort = string(ctx.Request.Header.Peek(ServicePort))
		props.RequestID = string(ctx.Request.Header.Peek(RayID))
		props.ForwardedFor = string(ctx.Request.Header.Peek(ForwardedFor))
		props.Host = GenHostName(ctx)
		props.DataCenter = string(ctx.Request.Header.Peek(DataCenter))
		props.Proto = string(ctx.Request.Header.Peek(Proto))
	}

	if page, exists := cfg.Pages[pageCode]; exists {
		props.Message = page.Message()
		props.Description = page.Description()
	} else if c, err := strconv.Atoi(pageCode); err == nil {
		if s := fasthttp.StatusMessage(c); s != "Unknown Status Code" { // as a fallback
			props.Message = s
		}
	}

	SetClientFormat(ctx, PlainTextContentType) // set default content type

	if props.Message == "" {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		_, _ = ctx.WriteString("requested pageCode (" + pageCode + ") not available")

		return
	}

	// proxy required HTTP headers from the request to the response
	for _, headerToProxy := range opt.ProxyHTTPHeaders {
		if reqHeader := ctx.Request.Header.Peek(headerToProxy); len(reqHeader) > 0 {
			ctx.Response.Header.SetBytesV(headerToProxy, reqHeader)
		}
	}

	switch {
	case clientWant == JSONContentType && canJSON: // JSON
		{
			SetClientFormat(ctx, JSONContentType)

			if content, err := rdr.Render(json.Content(), props); err == nil {
				ctx.SetStatusCode(httpCode)
				_, _ = ctx.Write(content)
			} else {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				_, _ = ctx.WriteString("cannot render JSON template: " + err.Error())
			}
		}

	case clientWant == XMLContentType && canXML: // XML
		{
			SetClientFormat(ctx, XMLContentType)

			if content, err := rdr.Render(xml.Content(), props); err == nil {
				ctx.SetStatusCode(httpCode)
				_, _ = ctx.Write(content)
			} else {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				_, _ = ctx.WriteString("cannot render XML template: " + err.Error())
			}
		}

	default: // HTML
		{
			SetClientFormat(ctx, HTMLContentType)

			var templateName = p.Pick()

			if template, exists := cfg.Template(templateName); exists {
				if content, err := rdr.Render(template.Content(), props); err == nil {
					ctx.SetStatusCode(httpCode)
					_, _ = ctx.Write(content)
				} else {
					ctx.SetStatusCode(fasthttp.StatusInternalServerError)
					_, _ = ctx.WriteString("cannot render HTML template: " + err.Error())
				}
			} else {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				_, _ = ctx.WriteString("template " + templateName + " not exists")
			}
		}
	}
}
