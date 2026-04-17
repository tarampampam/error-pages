# AGENTS - Project Rules

> Read this file AND the global rules before making any code changes -
> https://tarampampam.github.io/.github/ai/AGENTS.md (mirror -
> <https://raw.githubusercontent.com/tarampampam/.github/refs/heads/master/ai/AGENTS.md>).

## Instruction Priority

1. This file (`AGENTS.md` in this repository)
2. Global rules (external URLs)
3. Other documentation

If rules conflict, follow the highest priority source.

## Commands

```bash
# Build
make build # go generate + compile to ./error-pages (trimpath, strips debug symbols)
make gen   # go generate ./... - regenerates README.md from CLI definitions (run after changing flags)

# Test
make test # go test -race ./...

# Lint
make lint # golangci-lint run (must be installed separately)

# Build static pages
mkdir ./tmp
./error-pages build --target-dir ./tmp/pages --index
./error-pages build --target-dir ./tmp/pages --disable-l10n --disable-minification
```

## Module and language

- Module path: `gh.tarampamp.am/error-pages`
- Go 1.26 (see the `go.mod` file), FastHTTP (not `net/http`) for the HTTP server (due to performance reasons)
- Line length limit: **120 characters** (enforced by golangci-lint)

## Architecture overview

The project is a single-binary HTTP server and static page generator. All assets (HTML templates, l10n JS, favicon)
are embedded in the binary via `//go:embed`. There are no external runtime dependencies.

### Package layout

```
cmd/error-pages/main.go         - CLI entrypoint: signal context + cli.NewApp()
internal/
  appmeta/                      - version string (injected via -ldflags at build time)
  cli/
    app.go                      - root CLI command, registers subcommands, initialises logger
    serve/command.go            - "serve" subcommand: parses flags → builds Config → starts HTTP server
    build/command.go            - "build" subcommand: renders all templates×codes to disk
    shared/flags.go             - shared flag definitions (reused by serve and build)
  config/
    config.go                   - Config struct + defaults (templates, codes, formats, feature flags)
    templates.go                - templates map[name]content with CRUD + RandomName()
    codes.go                    - Codes map with wildcard-aware Find()
    rotation_mode.go            - RotationMode enum (disabled/random-on-startup/per-request/hourly/daily)
  http/
    server.go                   - FastHTTP server: routing, middleware wiring, graceful shutdown
    handlers/error_page/        - main handler: code extraction, format detection, rendering, caching
    handlers/live/              - GET /healthz → "OK\n"
    handlers/version/           - GET /version → {"version":"x.y.z"}
    handlers/static/            - GET /favicon.ico (embedded binary)
    middleware/                 - HTTP server middleware
  logger/                       - slog wrapper (console/JSON formats, named sub-loggers)
  template/
    template.go                 - Render(content, Props) using text/template + built-in FuncMap
    props.go                    - Props struct with `token:"..."` tags; Values() via reflection
    minify.go                   - MiniHTML() wrapping tdewolff/minify (HTML+CSS+SVG+JS)
templates/                      - built-in HTML themes, embedded via //go:embed *.html
l10n/                           - l10n.js (client-side JS, 16 locales), embedded via //go:embed
```

## Request pipeline

- fasthttp.Server.Handler (server.go; manual path switch - no router framework)
- error_page.New() handler (handlers/error_page/handler.go)
  * Code extraction (first match wins)
  * Determining HTTP status for response
  * Response format detection
  * Setting response headers
  * Proxy headers (cfg.ProxyHeaders; copied from request to response if present)
  * Template Props building
  * Template selection via templateToUse
  * Cache check
  * Render errors → inline error HTML/JSON/XML strings (never panics)

## Configuration system

`config.New()` initialises all defaults in Go code. No config files. The `serve` command applies flags on top;
flags read from env vars or CLI (CLI wins on conflict).

## Template system

### Built-in themes

`app-down`, `cats`, `connection`, `ghost`, `hacker-terminal`, `l7`, `lost-in-space`, `noise`, `orient`,
`shuffle`,`win98`.

All embedded in binary via `//go:embed *.html` in `templates/embed.go`. `cats` is the only theme fetching external
resources (cat image CDN).

### Template language

Templates use Go's `text/template` (not `html/template`). Each template receives a `template.Props` and a FuncMap.
Fields can be accessed both as `.FieldName` and as zero-argument functions (no dot):

```
{{ code }}           uint16  - HTTP status code
{{ message }}        string  - short status message ("Not Found")
{{ description }}    string  - longer description
{{ original_uri }}   string  - X-Original-URI header (ingress-nginx)
{{ namespace }}      string  - X-Namespace header (ingress-nginx)
{{ ingress_name }}   string  - X-Ingress-Name header (ingress-nginx)
{{ service_name }}   string  - X-Service-Name header (ingress-nginx)
{{ service_port }}   string  - X-Service-Port header (ingress-nginx)
{{ request_id }}     string  - X-Request-Id header
{{ forwarded_for }}  string  - X-Forwarded-For header
{{ host }}           string  - Host header
{{ show_details }}   bool    - true if --show-details is set (same as !hide_details)
{{ hide_details }}   bool    - inverted show_details (convenience)
{{ l10n_disabled }}  bool    - true if --disable-l10n is set
{{ l10n_enabled }}   bool    - inverted l10n_disabled (convenience)
```

Built-in utility functions:

```
{{ nowUnix }}                            - current Unix timestamp (int64)
{{ hostname }}                           - server hostname
{{ version }}                            - app version string
{{ env "VAR_NAME" }}                     - os.Getenv
{{ escape "<html>" }}                    - html.EscapeString
{{ json .value }}                        - json.Marshal → string (safe for any type)
{{ int .value }}                         - cast to int (returns 0 on failure)
{{ l10nScript }}                         - injects full l10n.js content as inline script
{{ strContains "haystack" "needle" }}
{{ strCount "string" "substr" }}
{{ strTrimSpace "  val  " }}
{{ strTrimPrefix "foobar" "foo" }}
{{ strTrimSuffix "foobar" "bar" }}
{{ strReplace "string" "old" "new" }}    - replaces ALL occurrences
{{ strIndex "foobar" "bar" }}            - returns index or -1
{{ strFields "a b c" }}                  - strings.Fields → []string
```

### How Props → FuncMap mapping works

`props.Values()` uses reflection to iterate `Props` struct fields, reading their `token:"..."` struct tag as the
map key. Each token becomes a zero-arg function in the FuncMap returning `any`. This means **adding a new template
variable requires**: (1) a new field in `Props` with a `token` tag, (2) filling it in `handler.go`, and (3) nothing
else - it auto-registers.

### Custom templates

Load at runtime with `--add-template /path/to/file.html`. Template name is derived from the file's basename without
extension. The file must be a valid Go `text/template` string. All built-in functions listed above are available.

## HTTP code wildcard matching

`config.Codes` is `map[string]CodeDescription` where keys can be wildcards:

- Exact: `"404"` - matches only 404
- Wildcard chars: `*`, `x`, `X` in any position - e.g. `"4xx"`, `"4**"`, `"4*X"`
- Length must match (all keys are 3 chars for standard HTTP codes)
- Specificity: fewer wildcards = more specific = wins if multiple patterns match
- Example: codes map has `"4xx"` and `"404"` → request for 404 gets `"404"` entry; request for 405 gets `"4xx"` entry

Wildcard codes are **valid in the config map** but the `build` command skips them (only numeric keys generate `.html`
files, since `strconv.ParseUint` is used).

## Response format detection

Priority chain in `detectPreferredFormatForClient()`:

1. `Content-Type` request header - parsed before `;`, e.g. `text/html; charset=utf-8` → html
2. `X-Format` request header - treated as raw accept string (ingress-nginx sends original Accept here)
3. `Accept` request header - parsed with q-factor weights; `*/*` is explicitly ignored; highest weight wins
4. Fallback - `unknownFormat` which renders as plain text

MIME type matching (case-insensitive, substring):

- `/json` → JSON (matches `application/json`, `text/json`)
- `/xml` or `+xml` → XML (matches `application/xml`, `application/xhtml+xml`)
- `/html` → HTML
- `/plain` → plain text

## Render cache

`RenderedCache` in `handlers/error_page/cache.go`:

- TTL: **900ms** - must stay below 1 second so `{{ nowUnix }}` remains accurate.
- Key: `[32]byte` = `MD5(template_string)[0:16]` + `MD5(gob.Encode(props))[16:32]`.
- Thread-safe: `sync.RWMutex`.
- Background goroutine fires every TTL to call `ClearExpired()`; stopped by closing a channel (returned
  as `closeCache func()` from `New()`).
- Minification result is what gets cached - minification runs before `cache.Put`.

## Localization (l10n)

Client-side only. `l10n.js` is injected verbatim into HTML via `{{ l10nScript }}`. Templates that want l10n support
must call this function and then set `data-l10n` attributes on DOM elements.

- **Only ISO 639-1 two-letter codes** - no BCP 47 regional variants
- Detection: `navigator.language` (browser locale), matched against the data map
- Translation keys are English strings, normalised by lowercasing and stripping non-alphanumeric chars
- Disabled with `--disable-l10n` / `DISABLE_L10N=true` (sets `L10nDisabled=true` in Props → `l10n_enabled` returns
  false → templates conditionally skip the script)

To add a new translation: add entries to `const data` in `l10n/l10n.js` with the new locale code. To add a new
language, add it to every existing string's Map.

## Docker image

Multi-stage Dockerfile:

1. **Build stage** (`golang:alpine`): compiles binary, then runs `./error-pages build --target-dir /opt/html` to
  pre-generate all static pages.
2. **Runtime stage** (`scratch`): copies binary + `/opt/html`; runs as UID 10001 (non-root).

## Adding a new feature - checklist

### New CLI flag (serve command)

1. Define in `internal/cli/shared/flags.go` if reusable, or inline in `serve/command.go`
2. Add field to `internal/config/config.go` → `Config` struct
3. Set default in `config.New()` if non-zero
4. Wire flag → config field in `serve/command.go` Action function
5. Wire in `build/command.go` if applicable to static generation
6. Run `make gen` to regenerate README

### New template variable

1. Add field to `internal/template/props.go` → `Props` struct with `token:"snake_case_name"` tag
2. Fill the field in `internal/http/handlers/error_page/handler.go` (Props construction block)
3. That's it - `props.Values()` via reflection auto-registers it as a template function

### New built-in template function (non-prop)

Add to `builtInFunctions` FuncMap in `internal/template/template.go`.

## Key constraints and gotchas

- **FastHTTP, not net/http** everywhere. Handler signature is `fasthttp.RequestHandler` = `func(*fasthttp.RequestCtx)`.
  Use `ctx.Request.Header.Peek("Name")`, `ctx.SetContentType(...)`, `ctx.Write(...)`.
- **`text/template` not `html/template`** - no automatic HTML escaping. Use `{{ escape .val }}` explicitly when
  inserting user-controlled strings into HTML templates.
- **Flags must not use pointers** in `shared/flags.go` - urfave/cli flag structs are copied into commands and have
  their own state; pointer sharing is not thread-safe (documented in the source comment).
- **`gob.Encode` is used for cache key hashing** - `Props` fields must be gob-encodable (all current types are). If
  you add a non-gob-encodable type, `hash()` will silently return `[16]byte{}`.
- **`RandomName()` uses map iteration** - deliberately exploits Go's randomized map iteration for "random" selection.
  It is not cryptographically random and distribution is uneven for small maps.
- **Cache TTL < 1s is mandatory** - `{{ nowUnix }}` calls `time.Now().Unix()`. If TTL ≥ 1s, a cached page would show
  a stale timestamp.
- **`--rotation-mode random-on-startup`** sets `cfg.TemplateName` in `serve/command.go` before `templateToUse()` is
  ever called. `templateToUse()` treats it identically to `disabled` - both return `cfg.TemplateName`.
- **Template name collision**: `--add-template` with a filename matching a built-in name overwrites the built-in.
  Use `--disable-template` to remove built-ins explicitly.
- **`make gen` must run after any CLI flag changes** - the README is auto-generated from CLI definitions. CI will
  catch this via the build job, but it's easy to forget locally.

## Offline Fallback Rules

> Apply these only if the external rule URLs above are inaccessible. The external rules are authoritative.

### Go

- Wrap errors with context: `fmt.Errorf("operation: %w", err)`. Return sentinel errors directly when they are unlikely.
- Use `xErr` naming when multiple errors are in scope (e.g. `readErr`, `writeErr`); use `if err := ...; err != nil`
  for single short-lived errors.
- Interfaces in the consumer package; keep them minimal; add `var _ Interface = (*Impl)(nil)` compile-time assertions.
- Exported declarations must have a doc comment starting with the identifier name, ending with a period.
- No `fmt.Print*` / `print` / `println`; no global variables; no `init()` without justification.
- Line length ≤ 120 characters.
- Test files: `package foo_test` (external); one `_test.go` per tested file; both outer and inner `t.Parallel()`;
  map-based table-driven tests with `give*` / `want*` keys.
