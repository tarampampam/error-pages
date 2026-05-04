# 💻 Command line interface

## HTTP server

<!--GENERATED:SERVER_CLI-->
```
Description:
   Start the HTTP server to serve the error pages

Usage:
   error-pages

Version:
   0.0.0@undefined

Options:
   --log-level="…"           Logging level (debug/info/warn/error) (default: info) [$LOG_LEVEL]
   --log-format="…"          Logging format (console/json) (default: console) [$LOG_FORMAT]
   --addr="…", --listen="…"  HTTP server address to listen on (IPv4 or IPv6) (default: 0.0.0.0) [$HTTP_ADDR, $LISTEN_ADDR, $ADDR]
   --port="…"                HTTP server TCP port number (default: 8080) [$HTTP_PORT, $LISTEN_PORT, $PORT]
   --default-error-page="…"  Default HTTP status code to render (default: 404) [$DEFAULT_ERROR_PAGE]
   --send-same-http-code     The HTTP response should use the same status code as the requested error page [$SEND_SAME_HTTP_CODE]
   --show-details            Show details about the request in the error page response (if supported by the template) [$SHOW_DETAILS]
   --proxy-headers="…"       HTTP headers listed here will be proxied from the original request to the error page response (comma/new-line separated list) (default: X-Request-Id,X-Trace-Id,X-Correlation-Id,X-Amzn-Trace-Id) [$PROXY_HTTP_HEADERS]
   --disable-built-in-codes  Disable the built-in descriptions for HTTP status codes [$DISABLE_BUILT_IN_CODES]
   --add-code="…"            Add or override HTTP status codes and their messages/descriptions (format: 'CODE=MESSAGE[|DESCRIPTION][||CODE=MESSAGE[|DESCRIPTION]...]'; CODE may contain wildcards like '4**'; separate multiple entries with '||', a newline, or a tab) [$ADD_CODE]
   --template-name="…"       Name of the built-in HTML template to use (app-down/cats/connection/ghost/hacker-terminal/l7/lost-in-space/noise/orient/shuffle/win98; ignored if a custom HTML template is set) (default: app-down) [$TEMPLATE_NAME, $HTML_TEMPLATE_NAME]
   --rotation-mode="…"       Mode for rotating built-in HTML templates (disabled/random-on-startup/random-on-each-request/random-hourly/random-daily; ignored if a custom HTML template is set) (default: disabled) [$ROTATION_MODE]
   --homepage-url="…"        Homepage URL to show as a link in error pages (e.g. https://app.example.com/home) (default: /) [$HOMEPAGE_URL]
   --add-link="…"            Add extra links to error pages (format: 'LABEL=URL[||LABEL=URL...]'; separate multiple entries with '||', a newline, or a tab) [$ADD_LINK]
   --html-template="…"       Custom HTML template for error page responses (template text/URL/file path) [$HTML_TEMPLATE, $TEMPLATE]
   --json-template="…"       Custom JSON template for error page responses (template text/URL/file path) [$JSON_TEMPLATE]
   --xml-template="…"        Custom XML template for error page responses (template text/URL/file path) [$XML_TEMPLATE]
   --plaintext-template="…"  Custom plain text template for error page responses (template text/URL/file path) [$TEXT_TEMPLATE, $PLAINTEXT_TEMPLATE]
   --disable-l10n            Disable localization of error pages (if the template supports localization) [$DISABLE_L10N]
   --help, -h                Show help
   --version, -v             Print the version
```
<!--/GENERATED:SERVER_CLI-->

### Quick start

```bash
# run on default port 8080
./error-pages

# or with Docker
docker run --rm -p '8080:8080/tcp' ghcr.io/tarampampam/error-pages:4
```

Test it with curl:

```bash
# plain text (default when no Accept header is set - curl-friendly)
curl http://127.0.0.1:8080/404

# request a specific error code via header (path becomes irrelevant)
curl -H 'X-Code: 503' http://127.0.0.1:8080/

# request JSON format via Accept header
curl -H 'Accept: application/json' http://127.0.0.1:8080/503

# request JSON format via path extension
curl http://127.0.0.1:8080/503.json

# HTML
curl -H 'Accept: text/html' http://127.0.0.1:8080/404
```

### Custom templates

The `--html-template`, `--json-template`, `--xml-template`, and `--plaintext-template` flags each accept one of:

- **A file path** - template is read from disk at startup
- **A URL** - template is fetched over HTTP(S) at startup
- **Inline template text** - the value itself is used as the template

```bash
# from a local file
error-pages --html-template /etc/error-pages/my.html

# fetched from a URL at startup
error-pages --html-template https://example.com/my-error-template.html

# inline - useful in Kubernetes env vars or Docker Compose
error-pages --html-template '<html><body><h1>{{ .StatusCode }}: {{ .Message }}</h1></body></html>'
```

When `--html-template` is set, `--template-name` and `--rotation-mode` are ignored.

### Adding custom HTTP status codes

Add or override HTTP status code descriptions. Format: `CODE=MESSAGE|DESCRIPTION`.

```bash
# single entry
error-pages --add-code "418=I'm a teapot|Short and stout"

# multiple entries separated by ||
error-pages --add-code "418=I'm a teapot|Short and stout||499=Client Closed Request|The client closed the connection"

# wildcards: covers all 4xx codes not explicitly defined
error-pages --add-code "4**=Client Error|Something went wrong on your end"
```

Via environment variable (newline-separated):

```bash
ADD_CODE="418=I'm a teapot|Short and stout
499=Client Closed Request|The client closed the connection"
```

### Adding extra links

Add custom, labeled links (e.g. status page, contact, policy) to be displayed on every error page. Format: `LABEL=URL`.

```bash
# multiple links separated by || or newlines
error-pages --add-link "Status Page=https://status.example.com||Contact=https://example.com/contact"
```

URLs may contain `=` signs - only the first `=` in each entry is used as the separator.

## Templates builder

<!--GENERATED:BUILDER_CLI-->
```
Description:
   Build the static error pages and place them in the specified directory. If no custom template is provided, the built-in one will be used.

Usage:
   builder

Version:
   0.0.0@undefined

Options:
   --index                              Create an index.html file with links to all generated error pages [$CREATE_INDEX]
   --out="…", --target-dir="…", -o="…"  Directory to place the built error pages (default: .) [$OUT_DIR]
   --disable-built-in-codes             Disable the built-in descriptions for HTTP status codes [$DISABLE_BUILT_IN_CODES]
   --add-code="…"                       Add or override HTTP status codes and their messages/descriptions (format: 'CODE=MESSAGE[|DESCRIPTION][||CODE=MESSAGE[|DESCRIPTION]...]'; CODE may contain wildcards like '4**'; separate multiple entries with '||', a newline, or a tab) [$ADD_CODE]
   --template="…"                       Custom template for error pages [$TEMPLATE]
   --disable-l10n                       Disable localization of error pages (if the template supports localization) [$DISABLE_L10N]
   --homepage-url="…"                   Homepage URL to show as a link in error pages (e.g. https://app.example.com/home) [$HOMEPAGE_URL]
   --add-link="…"                       Add extra links to error pages (format: 'LABEL=URL[||LABEL=URL...]'; separate multiple entries with '||', a newline, or a tab) [$ADD_LINK]
   --help, -h                           Show help
   --version, -v                        Print the version
```
<!--/GENERATED:BUILDER_CLI-->

### Quick start

```bash
# generate all built-in templates into ./error-pages/
mkdir ./error-pages
builder --out ./error-pages

# generate using a custom HTML template
builder --template /path/to/my.html --out ./error-pages

# generate using a custom HTML template from a URL
builder --template https://example.com/my-error-template.html --out ./error-pages

# also create an index.html with links to all generated pages
builder --out ./error-pages --index
```

### Output structure

**Without `--template`** - all built-in templates are rendered, each into its own subdirectory:

```
./error-pages/
├── app-down/
│   ├── 400.html
│   ├── 401.html
│   ├── 403.html
│   ├── 404.html
│   └── ...
├── cats/
│   └── ...
├── ghost/
│   └── ...
└── ... (one subdirectory per built-in template)
```

**With `--template`** - only the specified template is rendered, files go into the root of `--out`:

```
./error-pages/
├── 400.html
├── 401.html
├── 403.html
├── 404.html
└── ...
```

### Adding extra links

The `--add-link` flag works the same way as in the HTTP server - see [Adding extra links](#adding-extra-links) above.

```bash
builder --add-link "Status Page=https://status.example.com||Contact=https://example.com/contact" --out ./error-pages
```
