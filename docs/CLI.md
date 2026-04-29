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
   --html-template="…"       Custom HTML template for error page responses (template text/URL/file path) [$HTML_TEMPLATE, $TEMPLATE]
   --json-template="…"       Custom JSON template for error page responses (template text/URL/file path) [$JSON_TEMPLATE]
   --xml-template="…"        Custom XML template for error page responses (template text/URL/file path) [$XML_TEMPLATE]
   --plaintext-template="…"  Custom plain text template for error page responses (template text/URL/file path) [$TEXT_TEMPLATE, $PLAINTEXT_TEMPLATE]
   --disable-l10n            Disable localization of error pages (if the template supports localization) [$DISABLE_L10N]
   --help, -h                Show help
   --version, -v             Print the version
```
<!--/GENERATED:SERVER_CLI-->

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
   --help, -h                           Show help
   --version, -v                        Print the version
```
<!--/GENERATED:BUILDER_CLI-->
