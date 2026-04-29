# AGENTS - Project Rules

`error-pages` is a Go HTTP server and static page generator that replaces default HTTP error responses (4xx/5xx) with
themed HTML pages.

## Two binaries

The project ships two separate binaries (two separate `cmd/` entries, two Docker image variants):

- **`error-pages`** (`cmd/error-pages/`) - HTTP server for dynamic error page serving.
- **`builder`** (`cmd/builder/`) - static generator that pre-renders `{code}.html` / `{code}.json` / etc. files into
  an output directory.

Docker image tags:
- `X.Y.Z` / `X.Y` / `X` - contains the HTTP server (`error-pages` binary).
- `X.Y.Z-builder` / `X.Y-builder` / `X-builder` - contains the builder binary and pre-rendered static pages.

## Deployment targets

- **Kubernetes + ingress-nginx**: runs as `defaultBackend`; ingress forwards error responses via the `X-Code` header
  and the `custom-http-errors` config.
- **Traefik**: wired as errors middleware; Traefik rewrites error responses to `/{status}.html` on this service.
- **Static nginx image**: `builder` pre-generates `{code}.html` files, copied into a custom nginx Docker image via
  `COPY --from=ghcr.io/tarampampam/error-pages:<tag>-builder`.

## How the HTTP server works

Routes (no router/mux - plain `switch` on `r.URL.Path`):

| Path / condition                               | Handler            |
|------------------------------------------------|--------------------|
| `/healthz`, `/health`, `/health/live`, `/live` | liveness probe     |
| `/version`                                     | version info       |
| `/favicon.ico`                                 | built-in favicon   |
| everything else                                | error page handler |

The error page handler resolves the HTTP code and response format from the incoming request:
- **Code**: from `X-Code` header, path (e.g. `/404`, `/404.html`, `/404.json`), or falls back to `--default-error-page`.
- **Format**: from `X-Format` header, `Content-Type` header, `Accept` header, or path extension (`.html`, `.json`,
  `.xml`, `.txt`). Defaults to plain text (curl-friendly).

Supported formats: `HTML`, `JSON`, `XML`, `PlainText` - see `internal/formats/format.go`.

Response extras:
- gzip compression when client sends `Accept-Encoding: gzip`.
- `Retry-After: 120` header for 408, 425, 429, 500, 502, 503, 504 codes.
- `X-Robots-Tag: noindex, nofollow, nosnippet, noarchive` on all responses.
- Selected request headers proxied to the response (default: `X-Request-Id`, `X-Trace-Id`, `X-Correlation-Id`,
  `X-Amzn-Trace-Id`; configurable via `--proxy-headers` / `$PROXY_HTTP_HEADERS`).

## Template system

All templates are parsed **once at startup** (not per request).

### Built-in HTML templates

Located in `templates/html/*.tpl.html`. Available names:
`app-down`, `cats`, `connection`, `ghost`, `hacker-terminal`, `l7`, `lost-in-space`, `noise`, `orient`, `shuffle`, `win98`.

Selected via `--template-name` / `$TEMPLATE_NAME` (default: `app-down`).

Rotation modes (`--rotation-mode` / `$ROTATION_MODE`):
`disabled` (default), `random-on-startup`, `random-on-each-request`, `random-hourly`, `random-daily`.

### Custom templates

Set via flags / env vars. Source can be: a URL (fetched at startup), a file path, or literal template text.

| Flag                   | Env var                                 | Format     |
|------------------------|-----------------------------------------|------------|
| `--html-template`      | `$HTML_TEMPLATE`, `$TEMPLATE`           | HTML       |
| `--json-template`      | `$JSON_TEMPLATE`                        | JSON       |
| `--xml-template`       | `$XML_TEMPLATE`                         | XML        |
| `--plaintext-template` | `$TEXT_TEMPLATE`, `$PLAINTEXT_TEMPLATE` | plain text |

Default non-HTML templates live in `templates/default.tpl.{json,xml,txt}`.

When a custom HTML template is set, `--template-name` and `--rotation-mode` are ignored.

### Template data structure

```go
// internal/template/data.go
type Data struct {
    StatusCode   uint16
    Message      string  // short status text
    Description  string  // longer description
    OriginalURI  string  // (ingress-nginx, requires --show-details)
    Namespace    string
    IngressName  string
    ServiceName  string
    ServicePort  string
    RequestID    string
    ForwardedFor string
    Host         string
    Config       Config
}

type Config struct {
    ShowRequestDetails bool
    L10nDisabled       bool
}
```

### Template functions (v4)

Pipeline-friendly (needle before haystack). See `internal/template/functions.go` for the full list. Key ones:

`now` (returns `time.Time`), `hostname`, `version`, `env`, `toJson`/`toJSON`, `toInt`/`int`,
`escape`, `trim`, `trimPrefix`, `trimSuffix`/`trimPostfix`, `replace`, `lower`, `upper`,
`default`, `coalesce`, `ternary`, `contains`, `hasPrefix`, `hasSuffix`, `count`,
`split`, `join`, `fields`, `quote`, `squote`, `repeat`, `substr`, `truncate`, `trimAll`,
`urlEncode`, `toString`/`str`, `isEmpty`, `isNotEmpty`, `not`, `l10nScript`.

> **`env` security**: keys containing `PASSWORD`, `SECRET`, `KEY`, `TOKEN`, `PASS`, `PWD`, or `CRED`
> (case-insensitive, `_`-segment matched) return a masked `***` string.

Deprecated v3 aliases still work (`nowUnix`, `json`, `strContains`, etc.) but argument order differs - do not
blindly rename them without flipping args. The v3→v4 syntax shim (`convertV3toV4`) is deprecated and will be
removed eventually.

### Localization

Client-side only. `l10n/localize.js` is injected verbatim into HTML via `{{ l10nScript }}`.
Templates that want l10n must call this function and set `data-l10n` attributes on DOM elements.
See [l10n/readme.md](l10n/readme.md).

## HTTP codes

`internal/codes/` - built-in descriptions for standard HTTP codes plus wildcard support (`4xx`, `4**`, `4XX`).

Adding or overriding codes: `--add-code` / `$ADD_CODE`.
Format: `CODE=MESSAGE|DESCRIPTION`. Multiple entries separated by `||`, newline, or tab.

```bash
--add-code "418=I'm a teapot|Short and stout||499=Custom Error|Another description"
```

Disable all built-in descriptions: `--disable-built-in-codes` / `$DISABLE_BUILT_IN_CODES`.

## Key internal packages

| Package                                   | Purpose                                                |
|-------------------------------------------|--------------------------------------------------------|
| `internal/httpserver`                     | HTTP server setup, handler wiring, middleware chain    |
| `internal/httpserver/handlers/error_page` | core error page rendering handler                      |
| `internal/httpserver/handlers/live`       | liveness probe handler                                 |
| `internal/httpserver/handlers/version`    | version endpoint handler                               |
| `internal/httpserver/middleware`          | access log and request log injection middleware        |
| `internal/template`                       | template parsing, rendering, rotation, data types      |
| `internal/template/tploader`              | loads template content from URL, file, or literal text |
| `internal/formats`                        | Format enum + MIME types + fallback error formatting   |
| `internal/codes`                          | built-in HTTP code descriptions, wildcard lookup       |
| `internal/cli`                            | minimal CLI flag/command parsing (no external deps)    |
| `internal/logger`                         | structured logger                                      |
| `internal/appmeta`                        | version string                                         |
| `internal/errgroup`                       | errgroup helper                                        |
| `internal/testutil/assert`                | test assertion helpers                                 |
| `templates/`                              | embedded HTML and default non-HTML templates           |
| `l10n/`                                   | localization JS and strings                            |
| `deploy/helm/`                            | Helm chart for Kubernetes deployment                   |

## Module and language

- Module path: `gh.tarampamp.am/error-pages/v4`
- Go 1.26 (see `go.mod`); **standard library only** - `net/http` for the server, zero external runtime dependencies.
- Line length: **≤ 120 chars** (enforced by golangci-lint).

## Working principles

**Match the codebase, don't reshape it.**

- Before writing code in a package, read 1-2 files most analogous to your change. Use them as the style reference.
- Prefer existing patterns. Don't introduce new abstractions, packages, or APIs without explicit approval - you may
  suggest them, but do not implement them immediately.
- Match existing style even if you'd do it differently.
- Make minimal, surgical changes. Every changed line should trace to the user's request.
- Don't refactor or "improve" code outside the task. If you spot dead code or pre-existing bugs, mention them - don't
  fix them.
- When your edits orphan imports, vars, or functions, clean those up. Don't delete pre-existing dead code.
- No global state (loggers, configs) without clear precedent in the codebase.
- Don't suppress linter errors (`//nolint`) without strong justification.
- Don't modify generated files.

**Don't guess.**

- Don't assume APIs, functions, or types exist - verify against the codebase.
- If two or more reasonable implementations exist, present them; don't pick silently.
- If something is genuinely ambiguous, stop and ask one focused question. One good question beats a discarded
  implementation.
- Changes affecting generated files, data formats, or public APIs require explicit approval before you act.

## Workflow

For non-trivial tasks, transform the request into a verifiable goal before coding:

- "Add validation" → write tests for invalid inputs, then make them pass.
- "Fix the bug" → write a test that reproduces it, then make it pass.
- "Refactor X" → ensure tests pass before and after.

After modifying any file, run these steps in order before considering the task complete:

1. **Read analogous files** in the same package (one or two, not all of them) to lock in the local style.
2. **Lint**: run linters, fix all errors and warnings.
3. **Test**: run tests, fix failures.
4. **Self-review** the diff against this checklist:
  - [ ] Logic: off-by-one, wrong operator, inverted condition, unreachable branch.
  - [ ] Concurrency: missing locks, shared state without synchronization, deadlocks.
  - [ ] Errors: silently swallowed, missing checks, wrong sentinel comparison.
  - [ ] Security: unsanitized input in SQL/shell, secrets in code, weak auth checks.
  - [ ] Pre-existing bugs you stumbled into - report only, don't fix unprompted.
5. **Update [README.md](README.md)** if the change is user-facing: features, config/env/flags/defaults,
   CLI/API, deprecations, breaking changes. Skip for purely internal changes (refactors, tests, CI).
6. **Update [AGENTS.md](AGENTS.md)** if a future agent needs to know something new about this codebase.

Don't present work as finished until lint and tests pass cleanly. Don't fix issues outside scope without asking first.

## Go rules

### Errors

- Wrap with context when it helps debugging: `fmt.Errorf("operation: %w", err)`.
- Don't wrap when failure is improbable; return as-is: `if _, err := buf.Write(data); err != nil { return err }`.
- Define sentinel errors at package level when callers are expected to check them via `errors.Is`:
  `var ErrNotFound = errors.New("not found")`.

**Naming**: prefer `xErr` (e.g. `pingErr`, `decodeErr`, `wErr`) when multiple errors are in scope, to avoid
overwriting a single `err` and accidentally checking the wrong one. Plain `err` is fine for short-lived inline
`if err := ...; err != nil` blocks. Style convention, not a hard rule.

```go
ping, pingErr := some.Ping(ctx)
if pingErr != nil { ... }

n, wErr := buf.Write(b)
if wErr != nil { ... }
```

### Interfaces

- Define interfaces in the consumer package.
- Keep them minimal.
- Add compile-time assertions: `var _ Interface = (*Impl)(nil)`.

### Comments

**Doc comments** on exported declarations:

- Start with the identifier name (godoc convention), end with a period.
- Describe primary purpose, edge cases, and motivation for non-obvious behavior.
- Don't enumerate params or errors unless one is genuinely surprising or load-bearing for callers.
- Technical English only.
- Use a plain hyphen `-` as a separator. Never em dashes, never arrows (`←` / `→`).

```go
// Queue enqueues the item with the given ID for processing. It returns an error if the item cannot be enqueued due to
// transient issues (e.g. DB failure). The method is idempotent and safe to call multiple times with the same ID.
func (h *Handler) Queue(ctx context.Context, id string) error { ... }

// ErrNotFound is returned by DB query methods when the requested record does not exist.
var ErrNotFound = errors.New("not found")
```

**Inline comments** inside function bodies:

- Only when the code is genuinely non-obvious. Explain *why*, not *what*.
- Lowercase first letter, no trailing period.
- Same hyphen / em dash / arrow rule as above.

```go
// not found is a valid outcome here - the caller treats absence as permission to proceed
if errors.Is(err, db.ErrNotFound) {
  return nil
}
```

Don't comment obvious assignments, type conversions, stdlib calls, or anything the name already conveys.

### Linter rules

- No `fmt.Print*`, `print`, `println`.
- No global variables.
- No `init()` without justification.
- Line length ≤ 120.
- Import order: stdlib → external → internal.

See [.golangci.yml](.golangci.yml) for the full set.

## Testing

### Structure

- External test package: `package foo_test` (not `package foo`) - prevents accidental access to unexported identifiers.
- One `_test.go` file per source file (e.g. `utils.go` → `utils_test.go`).
- Both outer `t.Parallel()` and inner `t.Parallel()` (inside `t.Run`) are required.
- **Map-based table-driven tests** - maps give random ordering, which surfaces ordering-dependent bugs.
- Map key = test case name. Value = anonymous struct with `give*` (inputs) and `want*` / `checkErr` (expectations).
- Use `gh.tarampamp.am/error-pages/v4/internal/testutil/assert` for test assertions (`assert.NoError`,
  `assert.Equal`, etc.). No third-party assertion library.

### Principles

- Test behavior, not implementation.
- Cover happy path + key failures. Don't aim for 100% coverage. Don't over-test trivial code.
- Follow the existing test style in the project. Use the template below only when no clear pattern exists.

### When not to use a map

If the test exercises timing, channel behavior, or sequential logic that can't be cleanly expressed as independent
cases, use plain named `t.Run` subtests instead.
