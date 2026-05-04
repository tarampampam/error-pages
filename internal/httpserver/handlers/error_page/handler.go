package error_page

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"gh.tarampamp.am/error-pages/v4/internal/codes"
	"gh.tarampamp.am/error-pages/v4/internal/formats"
	"gh.tarampamp.am/error-pages/v4/internal/logger"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
)

// CodeDescriber is a function type that takes an HTTP status code and returns a description of that code, along
// with a boolean indicating whether the description was found.
type CodeDescriber func(uint16) (codes.Description, bool)

// Templater is a function type that takes a content format and returns a template for rendering error pages in
// that format.
type Templater func(formats.Format) (*tpl.Template, error)

// maxPooledBuf is the maximum buffer capacity that is returned to the pool to avoid retaining oversized allocations.
const maxPooledBuf = 64 << 10 // 64kb

// New creates a new handler that returns an error page with the specified status code and format.
func New( //nolint:funlen
	log *logger.Logger,
	defaultCode uint16,
	respondSameStatus bool,
	proxyHeaders []string,
	codeDescriber CodeDescriber,
	templater Templater,
	showDetails bool,
	l10nDisabled bool,
	homepageURL string,
	links []tpl.Link,
) http.Handler {
	// bufPool reuses the render buffer across requests to avoid per-request heap allocation for the response body
	bufPool := sync.Pool{New: func() any { return new(bytes.Buffer) }}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

			return
		}

		code, codeOk := getCodeFromRequest(r)
		if !codeOk {
			code = defaultCode
		}

		contentFormat, formatOk := getFormatFromRequest(r)
		if !formatOk {
			contentFormat = formats.PlainTextFormat // as default, curl-like clients prefer plain text over HTML
		}

		httpStatus := http.StatusOK
		if respondSameStatus {
			httpStatus = int(code)
		}

		for _, h := range proxyHeaders {
			if v := r.Header.Get(h); v != "" {
				w.Header().Set(h, v)
			}
		}

		w.Header().Set("Content-Type", contentFormat.ContentType())
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, nosnippet, noarchive")

		switch code {
		case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests,
			http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
			// tell the client (search crawler) to retry the request after 120 seconds
			w.Header().Set("Retry-After", "120")
		}

		codeDesc, descOk := codeDescriber(code)
		if !descOk {
			if std := http.StatusText(int(code)); std != "" {
				codeDesc.Short = std // use built-in HTTP status text as a fallback for unknown codes
			} else {
				codeDesc.Short = "Unknown Status Code" // ultimate fallback for non-standard codes
			}
		}

		tplData := tpl.Data{
			StatusCode:  code,
			Message:     codeDesc.Short,
			Description: codeDesc.Full,
			HomepageURL: homepageURL,
			Links:       links,
			Config: tpl.Config{
				ShowRequestDetails: showDetails,
				L10nDisabled:       l10nDisabled,
			},
		}

		//nolint:lll
		if showDetails { // ingress-nginx: https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/
			tplData.OriginalURI = r.Header.Get("X-Original-Uri")   // (ingress-nginx) URI that caused the error
			tplData.Namespace = r.Header.Get("X-Namespace")        // (ingress-nginx) namespace where the backend Service is located
			tplData.IngressName = r.Header.Get("X-Ingress-Name")   // (ingress-nginx) name of the Ingress where the backend is defined
			tplData.ServiceName = r.Header.Get("X-Service-Name")   // (ingress-nginx) name of the Service backing the backend
			tplData.ServicePort = r.Header.Get("X-Service-Port")   // (ingress-nginx) port number of the Service backing the backend
			tplData.RequestID = r.Header.Get("X-Request-Id")       // unique ID that identifies the request - same as for backend service
			tplData.ForwardedFor = r.Header.Get("X-Forwarded-For") // the value of the `X-Forwarded-For` header
			tplData.Host = r.Header.Get("Host")                    // the value of the `Host` header
		}

		buf, ok := bufPool.Get().(*bytes.Buffer)
		if !ok {
			buf = new(bytes.Buffer)
		}

		buf.Reset()

		tmpl, tErr := templater(contentFormat)
		if tErr != nil {
			buf.Write(contentFormat.FormatError("Failed to get the template for the requested content format: " + tErr.Error()))
		} else if tmpl == nil {
			buf.Write(contentFormat.FormatError("No template available for the requested content format"))
		} else if renderErr := tmpl.RenderTo(tplData, buf); renderErr != nil {
			buf.Write(contentFormat.FormatError("Failed to render the error page template: " + renderErr.Error()))
		}

		buf = gzipCompress(r, w, buf, &bufPool)

		w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		w.WriteHeader(httpStatus)

		if r.Method == http.MethodGet {
			if _, err := buf.WriteTo(w); err != nil {
				log.Error("Failed to write the response body", logger.Error(err))
			}
		}

		if buf.Cap() <= maxPooledBuf {
			bufPool.Put(buf)
		}
	})
}

// gzipCompress compresses src into a new buffer from pool if the request's Accept-Encoding header includes gzip.
// On success, it sets Content-Encoding and Vary response headers, returns src to pool, and returns the compressed
// buffer. Otherwise, it returns src unchanged.
func gzipCompress(r *http.Request, w http.ResponseWriter, src *bytes.Buffer, pool *sync.Pool) *bytes.Buffer {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || src.Len() == 0 {
		return src
	}

	dst, dstOk := pool.Get().(*bytes.Buffer)
	if !dstOk {
		dst = new(bytes.Buffer)
	}

	dst.Reset()

	gw := gzip.NewWriter(dst)

	_, wErr := gw.Write(src.Bytes())
	closeErr := gw.Close()

	if wErr != nil || closeErr != nil {
		if dst.Cap() <= maxPooledBuf {
			pool.Put(dst)
		}

		return src
	}

	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	if src.Cap() <= maxPooledBuf {
		pool.Put(src)
	}

	return dst
}
