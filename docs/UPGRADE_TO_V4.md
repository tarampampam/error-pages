# What's new in v4.0.0 and how to upgrade

## Intro

First off, I want to apologize for how rough this upgrade turned out. I didn't originally plan for it to be this
drastic, but somewhere along the way I realized that some of the decisions that looked fine at the start were
actually making the code messier and getting in the way of the project's growth.

So I made some tough calls. I'm convinced they're the right move long-term, but I know they'll cause some pain
during migration - which is exactly why this document exists, to help you through it.

If you're only here for the "what do I need to change" part (the upgrade guide), scroll down to the bottom of this doc.

## What's changed

### FastHTTP → net/http

I dropped the third-party HTTP library in favor of Go's standard library. This simplified the code, improved
compatibility, and got rid of some weird bugs along the way. FastHTTP is genuinely a really cool project - it
runs about 3x faster than stdlib - but for this project that level of performance isn't critical, and simplicity
and reliability matter more. On top of that, going back to stdlib brought HTTP/2 support out of the box
and lowered memory usage. Here's a side-by-side comparison of performance between the old and new versions:

![perf_infographic](https://habrastorage.org/webt/4a/2f/76/4a2f76f9397d16310fa1c3538cf02a5e.png)

The latency shown in the infographic above is the **average (avg)**, which can be misleading - a handful of slow
outliers pull the mean up and hide how fast the typical request actually is. Percentiles give a much more honest
picture - p50 is the median (half of all requests are faster), p90 and p99 show the tail behavior.

Measured on loopback (`127.0.0.1`), single connection, no artificial load
(`wrk -t1 -c1 -d5s --latency http://127.0.0.1:8080/...`, less is better):


|      Format |    p50     |  p90   | p99  (very rare - only 1% of responses are slower) |
|------------:|:----------:|:------:|:--------------------------------------------------:|
|        HTML | **121 µs** | 262 µs |                       420 µs                       |
|        JSON | **51 µs**  | 75 µs  |                       0.9 ms                       |
|         XML | **48 µs**  | 73 µs  |                       1.1 ms                       |
|  Plain text | **47 µs**  | 68 µs  |                       1.1 ms                       |
| HTML + gzip | **2.4 ms** | 3.1 ms |                       3.9 ms                       |
| JSON + gzip | **256 µs** | 510 µs |                       789 µs                       |

HTML responses are large (full rendered template, ~65 KB), which is why gzip compression takes noticeably more
time there. JSON/XML/text are compact structured responses, so they are fastest overall.

### New HTTP routes

I added the following routes that might come in handy for some use cases:

- `GET /{code}.{json|xml|txt}` - returns the error in the format matching the extension (e.g., `/404.json` returns
  a JSON response describing the 404 error). No more figuring out how to pass an `X-Format`/`Content-Type`/`Accept`
  header just to get the response format you want when you'd rather be explicit.
- All endpoints support `HEAD` requests and correctly set `Content-Length` in the response.

> [!WARNING]
> Content negotiation priority is now:
>
> - Path extension: `.html`, `.htm`, `.json`, `.xml`, `.txt`
> -`Content-Type` header
> -`X-Format` header
> -`Accept` header
> -Fallback to plain text

### All external dependencies removed

This was one of the main challenges of moving to stdlib, since a lot of functionality that used to come from
third-party libraries had to be reimplemented from scratch. Good news is none of it was rocket science, and I managed
to write everything I needed in about a week - but it still wasn't trivial.

Clean `go.mod`, clean conscience :D

### Templates (themes) reworked

Previously the app always operated on a "collection" of templates - there were built-in templates, and the user could
add their own that would supplement them.

Now the logic is simplified - there's a user template (which, by the way, can now be specified not just as a file
path but also directly as content, or even as a URL to a template that gets downloaded on startup). If it's set,
only that one is used. If it's not set, you can pick one of the built-in templates instead (template rotation mode
is still around too).

I think this is significantly easier to understand and use. Either yours, or one of the built-ins - no confusion.

On top of that, I improved how templates passed via environment variables or YAML manifests are handled. They now
support multi-line templates, which wasn't possible in the previous version. For example:

```yaml
env:
  - name: HTML_TEMPLATE
    # language=html
    value: |
      <html>
        <head><title>{{ .StatusCode }}: {{ .Message }}</title></head>
        <body>
          <h1>{{ .StatusCode }}: {{ .Message }}</h1>
          <p>{{ .Description }}</p>
        </body>
      </html>
  - name: JSON_TEMPLATE
    # language=json
    value: |
      {
        "error": true,
        "message": {{ .Message | toJson }}
      }
```

For adding your own error descriptions, you can still use `--add-code` (or `$ADD_CODE`), but now it's a single flag
instead of multiple, and the value format has changed:

```bash
--add-code "418=I'm a teapot/Short and stout"     # v3
--add-code "499=Custom Error/Another description" # v3
--add-code "418=I'm a teapot|Short and stout"     # v4 - single entry
--add-code "418=I'm a teapot|Short and stout||499=Custom Error|Another description" # v4 - multiple entries in one value
```

```yaml
env:
  - name: ADD_CODE # v4 - env var, newline-separated
    value: |
      418=I'm a teapot|Short and stout
      499=Custom Error|Another description
```

> This became obvious to me after reading [this discussion](https://github.com/tarampampam/error-pages/discussions/297),
> so thanks for leaving your thoughts in the repo - I genuinely appreciate it and I read all your ideas and suggestions,
> even when I don't always reply.

> [!WARNING]
> If you keep the old `/` syntax for `--add-code`, the parser won't error - your description will be silently glued onto
> the message. Grep your configs for `--add-code` or `$ADD_CODE` to find and fix them.

### Localization (l10n) got easier

Before, you had to open a JS file and add your translations there, which wasn't great. I extracted all localizable
strings into a separate JSON file, and the JS file is now generated from it at build time. Way simpler and more
reliable.

### All templates are now parsed once at startup

This used to happen on every request, which was inefficient - but it was a forced compromise to keep backward
compatibility with previous versions. As a consequence, "functions" and "data" in templates had to be separated, and
now you'll need to update all custom templates (if you used your own templates instead of the built-ins) to match the
new data structure passed to the template. But it's not as hard as it sounds, really.

On top of that, all the template functions have been completely reworked so they can be used in chained calls, the way
it's normally done in Go templates (e.g., in Helm and other projects). This significantly expanded their
capabilities - you can now do much more complex things without extra cognitive overhead.

I'll cover how to migrate your templates in more detail below, but the short version: I implemented support for the
old template format, so if you don't want to deal with updating them right now, you can do it later. The legacy
support won't last forever, but I'm not planning to remove it anytime soon, so don't stress.

### HTML minification replaced by full-response gzip compression

Previously HTML responses were minified, but the minification wasn't perfect and would occasionally break
templates - plus it didn't work for other formats (JSON, XML, plaintext). Now, instead of that, all responses can be
gzip-compressed (if the incoming request supports it), which is actually a more effective way to shrink the response
and works for all formats without exception.

### Docker image now contains only the HTTP server

The app is now shipped as 2 separate binaries - `error-pages` (HTTP server) and `builder` (a static generator that
produces pre-rendered static error pages). Adding new templates was bloating the docker image size, and
realistically you only need one of these two - either pre-rendered pages from `builder`, or the HTTP server for
dynamic page generation, but not both at the same time.

As for the docker image and its tags:

- The image name stays the same
- Tags in the `X.Y.Z` (and `X.Y`, `X`) format contain the HTTP server
- Tags with the `X.Y.Z-builder` (and `X.Y-builder`, `X-builder`) suffix contain the builder and pre-rendered error pages

The `serve` and `build` sub-commands no longer exist as such - they're separate binaries now.

Both image variants are signed with [cosign](https://github.com/sigstore/cosign) via keyless signing (GitHub OIDC),
so if you're running in an environment that verifies image signatures you can trust they came straight from the
release pipeline without any extra setup on your end.

### Helm chart

If you're running this in Kubernetes, there's now an official Helm chart. No need to write your own manifests from
scratch - the chart covers all the config options, handles the `ingress-nginx` defaultBackend and Traefik `errors`
middleware setups, and ships with every release.

Install it directly from the OCI registry:

```bash
helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages --version X.Y.Z
```

Or as a dependency in your own `Chart.yaml`:

```yaml
dependencies:
  - name: error-pages
    version: "X.Y.Z"
    repository: oci://ghcr.io/tarampampam/error-pages/charts
```

There's also a `helm-chart.tgz` attachment on every GitHub release if you'd rather install from a local file.

Every chart release is signed with [cosign](https://github.com/sigstore/cosign) using keyless signing via GitHub
OIDC, so you can verify the signature before deploying if supply chain security matters to you. The chart is also
registered on ArtifactHub, where you can browse versions, config options, and the changelog without leaving your
browser.

### Release artifacts

The list of files attached to each GitHub release has changed significantly. What used to be a handful of raw
platform binaries is now properly archived, covers both `error-pages` and `builder` across all supported platforms,
and includes a few extras.

Here's what you'll find in each release now:

```
error-pages-{os}-{arch}(.tar.gz|.zip)  - the HTTP server (linux + darwin + windows / amd64 + arm64)
builder-{os}-{arch}(.tar.gz|.zip)      - the static page builder (same platforms)
error-pages-static(.tar.gz|.zip)       - pre-rendered static pages (from the builder)
helm-chart.tgz                         - Helm chart
checksums.txt                          - SHA256 checksums for everything above
```

Previously the release only shipped raw unarchived binaries for the `error-pages` server, a single
`error-pages-static.zip`, and no checksums. If you have any scripts or CI pipelines that download release assets
by filename - make sure to update them.

> One concrete thing to watch out for - the old binary had no archive wrapper at all - the asset was literally named
> `error-pages-linux-amd64` (or `error-pages-linux-amd64.exe` on Windows). Now that same binary lives inside
> `error-pages-linux-amd64.tar.gz` / `error-pages-linux-amd64.zip`, so you'll need to extract it after downloading.

## 🔥 Upgrade guide

### Template syntax (Go templates)

Below are all the variables and functions available in templates, and their syntax. If you used your own templates,
you'll need to update them to match the new data structure passed to the template:

```
{{ code }}          →  {{ .StatusCode }}
{{ message }}       →  {{ .Message }}
{{ description }}   →  {{ .Description }}
{{ original_uri }}  →  {{ .OriginalURI }}
{{ host }}          →  {{ .Host }}
{{ request_id }}    →  {{ .RequestID }}
{{ namespace }}     →  {{ .Namespace }}
{{ ingress_name }}  →  {{ .IngressName }}
{{ service_name }}  →  {{ .ServiceName }}
{{ service_port }}  →  {{ .ServicePort }}
{{ forwarded_for }} →  {{ .ForwardedFor }}
{{ show_details }}  →  {{ .Config.ShowRequestDetails }}
{{ hide_details }}  →  {{ not .Config.ShowRequestDetails }}
{{ l10n_enabled }}  →  {{ not .Config.L10nDisabled }}
{{ l10n_disabled }} →  {{ .Config.L10nDisabled }}
```

> v4 still ships a temporary shim (`convertV3toV4`) that auto-rewrites the old syntax at parse time, so existing
> templates keep rendering - but the shim is **deprecated and will be removed someday**.

### CMD / args

The `serve`, `build`, and `healthcheck` subcommands are gone. The binary now runs the HTTP server directly:

```bash
error-pages serve --port 8080 # v3
error-pages --port 8080       # v4
```

Update your Dockerfile `CMD`, Kubernetes manifest `args`, systemd unit, or whatever else launches the binary. If you
relied on `CMD ["serve"]` in the official image, drop it.

### Renamed environment variables

Please review the new env var names and update them in your configs (some also changed format - like `ADD_CODE` - so
be extra careful when updating):

```
TEMPLATES_ROTATION_MODE   → ROTATION_MODE
RESPONSE_JSON_FORMAT      → JSON_TEMPLATE
RESPONSE_XML_FORMAT       → XML_TEMPLATE
RESPONSE_PLAINTEXT_FORMAT → TEXT_TEMPLATE
```

> The old names are **not** aliased - they will be silently ignored.

For the full list of new env vars, please refer to the [CLI documentation](CLI.md).

### Renamed flags

- `--json-format` → `--json-template`
- `--xml-format` → `--xml-template`
- `--plaintext-format` → `--plaintext-template`
- `--listen` is still accepted, but `--addr` is now the canonical name
- `--template-name` is still accepted, but `--template` and `--theme` were dropped

> [!WARNING]
> Some short flags like `-l`, `-p`, and `-t` are gone. Use the long forms.

### Template functions

**Good news:** all v3 function names still work as deprecated aliases - `nowUnix`, `json`, `strContains`, `strCount`,
`strTrimSpace`, `strTrimPrefix`, `strTrimSuffix`, `strReplace`, `strIndex`, `strFields`. Existing templates keep
rendering, but you should update them to the new names:

| v3 (deprecated)             | v4                                           |
|-----------------------------|----------------------------------------------|
| `nowUnix` (returns `int64`) | `now` (returns `time.Time`)                  |
| `json`                      | `toJson` / `toJSON`                          |
| `strContains`               | `contains`                                   |
| `strCount`                  | `count`                                      |
| `strTrimSpace`              | `trim`                                       |
| `strTrimPrefix`             | `trimPrefix`                                 |
| `strTrimSuffix`             | `trimSuffix` / `trimPostfix`                 |
| `strReplace`                | `replace`                                    |
| `strFields`                 | `fields`                                     |
| `strIndex`                  | *(no direct replacement, alias still works)* |

> [!IMPORTANT]
> **Argument order changed for the new names**. v3 string functions used Go's `strings.*` order (`haystack, needle`);
> the v4 names are pipeline-friendly (`needle, haystack`).
>
> ```
> {{ strContains "test" "es" }} ← v3 alias, still works
> {{ contains "es" "test" }}    ← v4, args flipped
> {{ "test" | contains "es" }}  ← v4, idiomatic pipeline form
> ```
>
> If you blind-replace `strContains` → `contains` (or any of the other `str*` functions) without flipping the
> arguments, your conditions will silently evaluate the wrong way around. Same applies to `strCount`, `strReplace`,
> `strTrimPrefix`, `strTrimSuffix`.

**Behavior changes worth knowing about:**

- `now` returns `time.Time`, not an int. To get the old Unix timestamp use `{{ now.Unix }}`. The old `nowUnix` alias
  still returns `int` for backward compatibility.
- `env` now masks values whose key contains `PASSWORD`, `SECRET`, `KEY`, `TOKEN`, `PASS`, `PWD`, or `CRED`
  (case-insensitive, segment-matched on `_`). The function returns a string of `*` of the same length instead of
  the actual value. If you were intentionally rendering one of these into a template, it won't work anymore.

**New functions** (non-exhaustive - see template docs for the full list):
`lower`, `upper`, `default`, `coalesce`, `ternary`, `hasPrefix`, `hasSuffix`, `split`, `join`, `quote`, `squote`,
`repeat`, `substr`, `truncate`, `trimAll`, `urlEncode`, `toString` / `str`, `isEmpty`, `isNotEmpty`. All designed for
pipeline use (`{{ .Message | default "Unknown" | upper }}`).

## Upgrade checklist

Run through this in order:

1. **Drop `serve`** from every place that launches the binary (Dockerfile `CMD`, K8s `args`/`command`, systemd
   `ExecStart`, Compose `command`, shell scripts, CI).
2. **Rename env vars** in your configs.
3. **Replace short flags**.
4. **Rename `--*-format` flags** to `--*-template` (`--json-format` → `--json-template`, etc.; if you used
   `--add-template` with one file → switch to `--html-template`)
5. **Rewrite `--add-code` values**: replace `/` with `|`. If you have many entries, consider collapsing them into
   one value with `||`.
6. **Custom templates** - if you used `--add-template` with one file → switch to `--html-template`.
7. **Update template syntax** in any custom templates you ship. The v3-shim currently makes them work, but it's
   deprecated.
8. **Update query path pattern** - remove `.html` from your error page URLs if you had it (e.g.,
   `/{status}.html` → `/{status}`) if you don't want to force HTML responses.
9. **Delete `--disable-minification` / `DISABLE_MINIFICATION`** from your config; it's a no-op now.
10. **Delete `--read-buffer-size` / `READ_BUFFER_SIZE`** from your config.
11. **Boot it up and test.** Hit `/404`, `/404.json`, `/404.xml`, `/404.txt` and confirm the right format comes back.

**Test everything before you deploy to production.**

---

If I missed something, or you've got questions about the upgrade - don't hesitate to open an issue or drop a note in
discussions, I'll happily help you out. And thanks for using this project, I really appreciate it, and I'm doing my
best to make it as good as it can be for you.
