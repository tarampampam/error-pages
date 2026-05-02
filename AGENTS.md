# AGENTS.md

Reference for AI coding agents working on this repository. Read this **before any code change**.

## Project

`error-pages` - Go HTTP server and static generator that replaces default 4xx/5xx error responses with
themed HTML pages.

- Module: `gh.tarampamp.am/error-pages/v4` - **Go >=1.26, stdlib only** (`go.mod` has no `require` block).
- HTTP/1.1 + HTTP/2 h2c on the same listener, no TLS.
- Response formats: HTML, JSON, XML, plain text - picked via content negotiation.
- Templates parsed once at startup (not per request). Gzip applied when `Accept-Encoding: gzip` is set.

## Two binaries

| Binary        | Entry              | Purpose                                                            |
|---------------|--------------------|--------------------------------------------------------------------|
| `error-pages` | `cmd/error-pages/` | HTTP server, dynamic rendering                                     |
| `builder`     | `cmd/builder/`     | Static generator, pre-renders `{code}.{html,json,xml,txt}` to disk |

Docker tags: `X.Y.Z` (server), `X.Y.Z-builder` (builder + pre-rendered pages at `/opt/html/`).

## Further docs

When this file doesn't cover what you need, [README.md](README.md) is the doc index. Notable links:

- `docs/templating.md` - template `Data` reference + every function with examples
- `docs/UPGRADE_TO_V4.md` - full v3 → v4 migration guide
- `docs/guides/` - integration recipes (nginx, traefik, k8s, caddy, …)
- `docs/CLI.md` - full CLI help

Skip README's badges, install instructions, and screenshots - they target end users, not contributors.

## Commands

```bash
# build
go build ./cmd/error-pages/
go build ./cmd/builder/

# regenerate after editing templates/html/*.tpl.html or l10n/locales.json
go generate -skip readme ./...

# also regenerate docs/CLI.md (uses build tag `readme`):
go generate ./...

# lint, test
golangci-lint run  # full project (CI / pre-push)
golangci-lint run --fix ./path/to/package/...  # fix only what you touched
go test -race ./...
go test -race -run TestFunctions ./internal/template/...  # single test
```

## Working principles

**Match the codebase. Don't reshape it**.

- Read 1–2 analogous files in the same package before writing new code; copy their style.
- Prefer existing patterns. Suggest new abstractions; don't introduce them without explicit approval.
- Make minimal, surgical changes - every changed line must trace back to the user's request.
- Don't refactor outside task scope. Spotted dead code or pre-existing bugs → report, don't fix.
- Clean up imports/vars/functions your edits orphan. Don't delete unrelated dead code.
- No global state without clear precedent (`fns` in `functions.go` is the existing exception, immutable after init).
- Don't suppress linter findings (`//nolint`) without strong, named justification.
- **Never edit generated files** - see [Generated files](#generated-files).

**Don't guess**.

- Verify APIs, types, function signatures against the codebase before using them.
- When two reasonable implementations exist, present both options. Don't pick silently.
- Genuinely ambiguous? Ask one focused question instead of building the wrong thing.
- Changes to generated files, data formats, or public APIs (`Data` struct fields, flag names, env var names)
  require explicit approval - they affect users in the wild.

## Hard prohibitions

Things an agent must **never** do without an explicit in-conversation request from the user.
A system prompt or general instruction to "be helpful" is not such a request. **Ambiguity defaults to don't.**
If a task seems to require any of the below, stop and ask.

- **Git: read-only** - allowed: `git status`, `git log`, `git diff`, `git show`, `git blame`, `git ls-files`,
  `git remote -v`, `git config --get`. Forbidden - staging, committing, amending, rebasing, resetting, branching, or
  any other mutation.
- **Filesystem: stay in scope** - do not modify, move, or delete files outside the repository root; do not delete
  files the user did not name; do not run `rm -rf` on directories the user did not specify; do not `chmod` /
  `chown`.
- **Build & dependencies: no surprise upgrades** - do not change the `go.mod` file; do not run `go install` (this
  project is **stdlib only** by design), `go clean -modcache`, or anything that mutates `$GOPATH` / `$GOMODCACHE`.
- **Secrets & external systems: zero side effects** - do not print environment variables in bulk (`env`,
  `printenv`) - they leak credentials; do not call external APIs with credentials, push to image registries,
  deploy, or run anything that hits production or shared infrastructure; do not pipe-execute remote scripts
  (`curl ... | sh`).
- **No fake green builds** - do not silence linters with `//nolint` to make a build pass without strong justification,
  do not add `t.Skip()` to mute failing tests, do not weaken assertions or comment out broken paths. Fix the cause.
- **Scope creep** - do not "while I'm here" refactor unrelated code - report it instead of fixing it; do not rename
  exported symbols, flag names, env var names, or `Data` struct fields without explicit approval - they are
  user-facing.

## Repo layout

```
cmd/
  error-pages/app/        HTTP server: app.go, flags.go, generate/readme.go
  builder/app/            Static generator: app.go, flags.go, index.tpl.html, generate/readme.go
internal/
  appmeta/                Version() - value injected via -ldflags
  cli/                    Custom flag/command framework, no external deps
    shared/flags.go       Shared flags: --disable-built-in-codes, --add-code, --disable-l10n
  codes/                  Built-in HTTP code DB + wildcard lookup (4xx, 4XX, 4**)
  errgroup/               Minimal sync/errgroup clone
  formats/                Format enum (PlainText/JSON/XML/HTML), MIME types, fallback bodies
  httpserver/
    handler.go            Plain switch on r.URL.Path - no mux
    server.go             HTTP server itself + Graceful shutdown (5s drain)
    handlers/             error_page (core), favicon, live, version
    middleware/           InjectLog, AccessLog, Apply chain
  logger/                 slog-based, console + JSON handler
  template/
    data.go               Data + Config structs (template input)
    functions.go          FuncMap (var fns)
    template.go           New(), RenderTo(), Render()
    templates.go          Multi-format manager + rotation modes
    convert.go            v3→v4 syntax shim (deprecated)
    tploader/             URL / file / inline source detection and loading
  testutil/assert/        Test assertions (assert.NoError, assert.Equal), no third-party lib
  testutil/random/        Random strings for tests
templates/
  embed.go                go:embed default.tpl.{json,xml,txt}
  embed_html.go           GENERATED - embeds html/*.tpl.html
  html/                   Built-in HTML templates
  generate/embed_html.go  Generator source
l10n/
  locales.json            Source of truth for translations
  embed.go                go:embed localize.min.js → l10n.L10n()
  localize.js, localize.min.js, playground.html  GENERATED
  generate/localize.go    Generator source
docs/
  CLI.md                  Contains partially generated docs (build tag: readme)
  templating.md, UPGRADE_TO_V4.md, guides/
deploy/helm/              Helm chart sources
```

## Generated files

**Do not edit by hand. Regenerate via `go generate`.**

| File                                                               | Generated by                                      |
|--------------------------------------------------------------------|---------------------------------------------------|
| `templates/embed_html.go`                                          | `go generate ./templates/...`                     |
| `l10n/localize.js`, `l10n/localize.min.js`, `l10n/playground.html` | `go generate ./l10n/...`                          |
| `docs/CLI.md`                                                      | `go generate ./...` (requires `readme` build tag) |

`//go:generate` directives are used in: `templates/embed.go`, `l10n/embed.go`, both `cmd/*/app/app.go`.

## HTTP server

### Routing

Plain `switch r.URL.Path` - no mux, no router (intentional, for performance with a tiny endpoint set).

| Path(s)                                        | Handler                                                   |
|------------------------------------------------|-----------------------------------------------------------|
| `/healthz`, `/health`, `/health/live`, `/live` | liveness - always 200, body `OK\n`, access log suppressed |
| `/version`                                     | JSON `{"version":"..."}`, GET/HEAD only                   |
| `/favicon.ico`                                 | embedded ICO                                              |
| anything else                                  | `internal/httpserver/handlers/error_page/` handler        |

Middleware chain: `InjectLog` → `AccessLog` → handler.

### Code resolution (first match wins)

1. URL path: first segment with extension stripped - `/404`, `/503.json`, `/404.html` → numeric part. Range 1–999.
2. `X-Code` header (1–3 chars, range 1–999).
3. `--default-error-page` flag (default `404`).

### Format resolution (first match wins)

1. URL extension: `.html`/`.htm` → HTML, `.json`, `.xml`, `.txt` (case-insensitive).
2. `Content-Type` header (part before `;`).
3. `X-Format` header (ingress-nginx forwards client `Accept` here).
4. `Accept` header - highest q-weight wins, `*/*` is ignored, no q means weight 10 (q=1.0).
5. Default: **plain text** (curl-friendly).

### Response invariants

- `X-Robots-Tag: noindex, nofollow, nosnippet, noarchive` on every response.
- `Retry-After: 120` only for **408, 425, 429, 500, 502, 503, 504**.
- Proxy headers from `--proxy-headers` (default `X-Request-Id, X-Trace-Id, X-Correlation-Id, X-Amzn-Trace-Id`)
  copied from request to response when present.
- HTTP status code: **always 200** by default. `--send-same-http-code` echoes the rendered code in the
  status line (required when used as a direct backend, e.g. ingress-nginx `defaultBackend`).
- Gzip: unbounded `sync.Pool` of `*bytes.Buffer`. Buffers with `Cap() > 64 KB` are **not returned** to
  the pool (GC'd) - same pool reused for render and gzip destination.

## Template system

### `tpl.Data` struct

Defined in the [data.go](internal/template/data.go) file.

**Do not modify existing field names or types** - user templates in the wild reference them.

### Template functions

Defined in the [functions.go](internal/template/functions.go) file (read this file to understand the available
functions and their behavior).

48 keys total - 39 active + 9 deprecated v3 aliases. Active set:

`now`, `hostname`, `version`, `env`, `toJson`/`toJSON`, `toInt`/`int`, `toString`/`str`,
`escape`, `urlEncode`, `trim`, `trimPrefix`, `trimSuffix`/`trimPostfix`, `trimAll`, `replace`,
`lower`, `upper`, `default`, `coalesce`, `ternary`, `contains`, `hasPrefix`, `hasSuffix`/`hasPostfix`,
`count`, `split`, `join`, `fields`, `quote`, `squote`, `repeat`, `substr`, `truncate`,
`isEmpty`, `isNotEmpty`, `l10nScript`.

**v4 pipeline order: needle before haystack**. `{{ "test" | contains "es" }}` → `contains(needle, haystack)`.

**`env` masking**: `getEnv` splits the key on `_`, uppercases segments, and matches against
`PASSWORD, SECRET, KEY, TOKEN, PASS, PWD, CRED`. If **any** segment matches, value becomes `*` repeated
to the original rune length.

### Deprecated v3 aliases - argument order is FLIPPED

| Alias           | Replacement      | Args flipped?                                  |
|-----------------|------------------|------------------------------------------------|
| `nowUnix`       | `now.Unix`       | -                                              |
| `json`          | `toJson`         | no                                             |
| `strCount`      | `count`          | **yes** (haystack, needle vs needle, haystack) |
| `strContains`   | `contains`       | **yes**                                        |
| `strTrimSpace`  | `trim`           | no                                             |
| `strTrimPrefix` | `trimPrefix`     | **yes**                                        |
| `strTrimSuffix` | `trimSuffix`     | **yes**                                        |
| `strReplace`    | `replace`        | **yes**                                        |
| `strIndex`      | (no replacement) | -                                              |
| `strFields`     | `fields`         | no                                             |

**Trap**: these aliases delegate to the stdlib funcs (e.g. `strings.Count` directly), which use
`(haystack, needle)` order. The v4 names use `(needle, haystack)`. Do not rename without flipping args.

### v3 → v4 token shim

Source code: [convert.go](internal/template/convert.go).

Auto-rewrites `{{ code }}` → `{{ .StatusCode }}`, `{{ show_details }}` → `{{ .Config.ShowRequestDetails }}`,
etc. at parse time. Two regexes (action block + identifier) and two lookup maps (`v3tov4Fields`,
`v3tov4Tokens`, 13 entries). Deprecated - will be removed once users migrate.

### Rotation modes

`disabled` (default), `random-on-startup`, `random-on-each-request`, `random-hourly`, `random-daily`.

Implemented with `atomic.Pointer[time.Time]` + `atomic.Pointer[string]` - no mutex.

**Rotation has no effect when `--html-template` is set**. Same for `--template-name`.

### Custom template loading (`tploader.LoadTemplateContent`)

Tries in order: HTTP/HTTPS URL (30s timeout, 5 MB cap) → existing file path (5 MB cap) → treat as inline
literal. All custom templates are loaded concurrently at startup via `errgroup`.

## Built-in HTML templates

`app-down` (default), `cats`, `connection`, `ghost`, `hacker-terminal`, `l7`, `lost-in-space`, `noise`,
`orient`, `shuffle`, `win98`, etc. Source: `templates/html/{name}.tpl.html`.

`cats` fetches images externally; the rest are self-contained.

## Built-in HTTP codes

400, 401, 403, 404, 405, 407, 408, 409, 410, 411, 412, 413, 416, 418, 429, 500, 502, 503, 504, 505 (can be extended
in future).

`Codes.Find(code)` resolution: exact 3-digit match → wildcard (`4xx`/`4XX`/`4**`, fewest wildcards wins).

Override or extend: `--add-code "CODE=MESSAGE|DESCRIPTION"`. Multiple entries via `||`, newline, or tab.
Disable all built-ins: `--disable-built-in-codes`.

## CLI flags

CLI framework: `internal/cli`. `Flag[T]` is generic over `bool | int | int64 | string | uint | uint64 | float64 | time.Duration`.

**Value precedence: Default → Env var → CLI flag (CLI wins)**.

Actual CLI flags and supported env vars are described in the [docs/CLI.md](docs/CLI.md) file, which is **partially**
generated from the sources.

## Common change patterns

### Extending `tpl.Data` with a new field

- Add the field to the `Data` struct. **Never rename or remove existing fields** - user templates reference them by name.
- Ensure template tests reflect the new field and the change is covered.
- Update `docs/templating.md` field table.
- Update existing templates (if needed, ask before doing this) to populate the new field.

### Adding a new CLI flag

- **CLI flag** (both binaries) → `internal/cli/shared`, with a test. Single-binary → that binary's own flags file.
- **Each binary's app wiring** → `opt` struct field, flag instantiation, `setIfFlagIsSet`, pass downstream,
  startup log line, etc.
- **Helm chart** → values file, deployment template (env var block), values JSON schema.
- **Update documentation** - README.md, `docs/CLI.md` (generated), other files in `docs/**/*.md` if relevant.

## Localization

Client-side only. `l10n/locales.json` (source of truth) → `localize.min.js` (embedded via
`l10n.L10n() string`) → injected into HTML via `{{ l10nScript }}` (renders inline `<script>…</script>`).

Browser resolves `navigator.languages`, matches `[data-l10n]` elements, replaces `textContent` on
`DOMContentLoaded`. BCP 47 fallback: `zh-TW → zh-tw → zh → en`. English is the passthrough.
Public JS API: `window.l10n.setLocale()`, `l10n.translate()`, `l10n.localizeDocument()`. 15+ languages.

Add a language: edit `locales.json`, run `go generate ./l10n/...`, verify in `playground.html`.

## Go rules

### Errors

- Wrap errors with context when it adds value for debugging - e.g. `fmt.Errorf("operation: %w", err)`.
  Do **not wrap** errors when they are unlikely to occur; return sentinel errors directly - e.g.
  `if _, err := buf.Write(data); err != nil { return err }`.
- Define sentinel errors at package level: `var ErrNotFound = errors.New("not found")` when they are expected to be
  checked by callers. Otherwise, return wrapped errors with context.
- Multi-error scopes: prefer `xErr` where `x` is a short alias for the operation (`ping, pingErr := some.Ping(ctx)`).
  Plain `err` is fine in short `if err := DoSomething(ctx); err != nil { ... }` blocks.

### Interfaces

- Define interfaces in the **consumer** package. Keep them minimal.
- Add compile-time assertion: `var _ Interface = (*Impl)(nil)`.
- `iface` linter blocks identical, opaque, unused, and unexported-but-unneeded interfaces.

### Receivers, type assertions, conversions

- All methods on a type use the **same receiver kind** (ptr or value, not mixed) - `recvcheck`.
- Type assertions must be two-value: `v, ok := x.(T)` - `forcetypeassert`.
- No unnecessary type conversions - `unconvert`.

### Comments

Doc comments on exported symbols (enforced by `godoclint` with `require-doc`):

- Start with the identifier name.
- End with a period (`godot`).
- Technical English only. Hyphen `-` as separator - never em dashes or arrows.

Inline comments inside function bodies:

- Only when the code is genuinely non-obvious. Explain *why*, not *what*.
- Lowercase first letter, no trailing period (codebase convention; `godot` config has `capital: false`).
- Same hyphen rule.

## Linter rules (golangci-lint)

Critical or non-obvious enforcement (full set in [.golangci.yml](.golangci.yml)):

- **No `fmt.Print*`, `print`, `println`** - `forbidigo` blocks debug prints.
- **No package-level `var`** where applicable (existing exception: `tpl.fns` FuncMap) - `gochecknoglobals`.
- **No `init()`** - `gochecknoinits`.
- **Line length: hard ceiling 120**. Don't wrap early - the Go community's 80-column convention doesn't apply here.
  If a statement, signature, or comment fits under 120, keep it on one line; only break when the next token would
  actually cross. Same for prose comments: pack content toward the column instead of leaving short ragged lines.
  `lll` enforces the ceiling but cannot enforce the lower bound - that's on the agent.
- **Import order**: stdlib → external → `gh.tarampamp.am/error-pages` - `gci` (config: `[standard, default, prefix(...)]`).
- **`//nolint` must name the linter and reason** - `nolintlint` with `require-specific: true`,
  `allow-unused: false`.
- **Function size**: 100 lines, 60 statements - `funlen`.
- **Cognitive/cyclomatic complexity**: 40 each - `gocognit`, `gocyclo`. Nested-if depth: 10 - `nestif`.
- **Repeated string literal ≥ 4 occurrences must be a const** - `goconst`.
- **Outbound calls must carry a context** - `noctx`.
- **`net/http` header keys must be canonical** - `canonicalheader`.
- **`testpackage`** + **`tparallel`** + **`paralleltest`** with `ignore-missing-subtests: true` -
  see [Testing](#testing).

## Testing

### Structure

- External package: `package foo_test` - required by `testpackage`.
- One `_test.go` per source file.
- **`t.Parallel()` required at the top level of every test**. Subtests should call `t.Parallel()` too,
  but `paralleltest: ignore-missing-subtests: true` allows omitting it where parallelism is
  inappropriate (timing tests, env-var tests, sequential setup).
- **Map-based table-driven tests.** Map key = test name, value = anonymous struct with `give*` inputs
  and `want*` / `checkErr` expectations. Random map iteration surfaces ordering-dependent bugs.
- **Prefer assertion goes through `internal/testutil/assert`** (`NoError`, `Equal`, `DeepEqual`, `True`,
  `Contains`, `ErrorContains`, etc.). Avoid `if got != want { t.Errorf(...) }`, no third-party library.
  Need an assertion that doesn't exist yet? Add it to the package.
- `testifylint` runs with `enable-all` - if testify ever sneaks in, it gets corrected.

### When the Map-based table-driven pattern doesn't fit

Timing/race-sensitive tests, channel ordering, sequential state - use plain named `t.Run` subtests
instead of a map.

### Principles

- Test behavior, not implementation.
- Cover happy path + key failure modes. Don't chase 100% coverage.
- Use `t.Setenv`, `t.TempDir`, `t.Context` (`usetesting` linter).

## Agent workflow (after any file change)

1. **Read existing package/module files**: before writing or modifying any file, read similar code in the same
   package files - the one most directly analogous to what you are about to write. One or two files is sufficient;
   do not read all files in the package. Use them as the authoritative style reference for that package.
2. If `templates/html/` or `l10n/locales.json` changed → `go generate -skip readme ./...`
3. **Lint:** run with `--fix` scoped to the packages you changed, e.g. `golangci-lint run --fix ./path/to/package/...`.
   `--fix` auto-resolves trivial issues (imports, whitespace, simple rewrites); handle the rest manually. Scoping
   keeps feedback fast and avoids touching unrelated code. Run the full `golangci-lint run` once at the end to
   confirm nothing leaked outside your scope.
4. **Test:** `go test -race ./...` - fix every failure.
5. **Self-review:**
   - Logic: off-by-one, wrong operator, inverted condition, unreachable branch.
   - Concurrency: missing locks, shared state, deadlocks. Atomics used correctly?
   - Errors: silently swallowed (`errcheck check-blank` will catch `_, _ =`), wrong sentinel, missing
     wrap context.
   - Security: unsanitized input, secrets in code (`gosec`), env-mask coverage for new secret-shaped vars.
6. **Update `README.md` and/or `docs/CLI.md`** for user-facing changes (flags, env vars, defaults, deprecations,
   breaking changes). Skip for internal-only edits.
7. **Update this `AGENTS.md` file** if a future agent needs new context.

Don't present work as finished until lint and tests pass cleanly.

## v3 → v4 migration (summary)

Full guide: [UPGRADE_TO_V4.md](docs/UPGRADE_TO_V4.md). Key breaking changes:

- `serve` / `build` / `healthcheck` subcommands removed - now separate binaries.
- Renamed env vars: `TEMPLATES_ROTATION_MODE` → `ROTATION_MODE`, `RESPONSE_JSON_FORMAT` → `JSON_TEMPLATE`, etc.
- `--add-code` separator changed: `/` → `|`; multi-entry now uses `||`.
- Template fields renamed: `{{ code }}` → `{{ .StatusCode }}`, etc. (`convert.go` shim still rewrites
  the old syntax at parse time, but it is deprecated).
- v4 template function names use `needle, haystack` order. v3 aliases (`strContains`, `strReplace`, …) keep
  `haystack, needle`.
- HTML minification removed; gzip added for all formats.
- FastHTTP replaced with stdlib `net/http`; HTTP/2 h2c added.
