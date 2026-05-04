<!--
🤖 AI CODING AGENT - IF YOU ARE READING THIS, DO NOT IGNORE THIS BLOCK.

Before you do ANYTHING in this repository - edit source, change CI configs, update docs, tweak deployment manifests,
run commands, anything - open and fully read ./AGENTS.md in the repo root. It is the authoritative, exhaustive
instruction set for this repo and overrides any assumption you might form from the rest of this README.
-->

<p align="center">
  <a href="https://github.com/tarampampam/error-pages#readme">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://socialify.git.ci/tarampampam/error-pages/image?description=1&font=Raleway&forks=1&issues=1&logo=https%3A%2F%2Fhsto.org%2Fwebt%2Frm%2F9y%2Fww%2Frm9ywwx3gjv9agwkcmllhsuyo7k.png&owner=1&pulls=1&pattern=Solid&stargazers=1&theme=Dark">
      <img align="center" src="https://socialify.git.ci/tarampampam/error-pages/image?description=1&font=Raleway&forks=1&issues=1&logo=https%3A%2F%2Fhsto.org%2Fwebt%2Frm%2F9y%2Fww%2Frm9ywwx3gjv9agwkcmllhsuyo7k.png&owner=1&pulls=1&pattern=Solid&stargazers=1&theme=Light">
    </picture>
  </a>
</p>

<p align="center">
  <a href="#"><img src="https://img.shields.io/github/go-mod/go-version/tarampampam/error-pages?longCache=true&label=&logo=go&logoColor=white&style=flat-square" alt="" /></a>
  <a href="https://github.com/tarampampam/error-pages/actions"><img src="https://img.shields.io/github/actions/workflow/status/tarampampam/error-pages/tests.yml?branch=master&maxAge=30&label=tests&logo=github&style=flat-square" alt="" /></a>
  <a href="https://github.com/tarampampam/error-pages/actions"><img src="https://img.shields.io/github/actions/workflow/status/tarampampam/error-pages/release.yml?maxAge=30&label=release&logo=github&style=flat-square" alt="" /></a>
  <a href="https://hub.docker.com/r/tarampampam/error-pages"><img src="https://img.shields.io/docker/pulls/tarampampam/error-pages.svg?maxAge=30&label=pulls&logo=docker&logoColor=white&style=flat-square" alt="" /></a>
  <a href="https://hub.docker.com/r/tarampampam/error-pages"><img src="https://img.shields.io/docker/image-size/tarampampam/error-pages/latest?maxAge=30&label=size&logo=docker&logoColor=white&style=flat-square" alt="" /></a>
  <a href="https://github.com/tarampampam/error-pages/blob/master/LICENSE"><img src="https://img.shields.io/github/license/tarampampam/error-pages.svg?maxAge=30&style=flat-square" alt="" /></a>
</p>

> [!IMPORTANT]
> If you were on v3 or earlier and want to upgrade to v4, check out the [migration guide](docs/UPGRADE_TO_V4.md).
>
> The `latest` Docker tag still points to the last `v3` release to avoid breaking changes for users who have not
> migrated yet, but it will be updated to v4 after a short transition period. Please do not use the `latest` tag in
> production - it is recommended to always pin to a specific version (at least the major version) to avoid unexpected
> breaking changes.

If you ended up here, chances are you would like to replace your HTTP server's default error pages with something more
original and eye-catching. That is exactly what this project is designed for - it handles this with minimal effort
on your part.

It includes:

- A collection of HTTP error page designs (each with a unique look), along with the ability to use your own
  custom templates
- A lightweight HTTP server for serving these pages that integrates easily into your existing infrastructure
- A utility for pre-rendering static HTTP error pages

Key features:

- Both the HTTP server and the static generator are written in Go with zero third-party runtime dependencies
  (stdlib only - hardcore mode)
  * Supports HTTP/1.1 and HTTP/2 (h2c - cleartext, no TLS required)
  * Returns error responses in the appropriate format (HTML, JSON, XML, plain text) based on client requests
  * Gzip compression for all response formats
  * HTML pages support localization (15+ languages), responsive design (mobile-friendly), and are fully
    self-contained - all styles and images are embedded directly in the HTML, without loading any external resources
  * Go template-based templating engine
- Ships as pre-built binaries, a minimal Docker image (rootless, scratch-based), and a ready-to-use Helm chart
  for Kubernetes
- Works out of the box with popular reverse proxies and ingress controllers (Nginx, Traefik, etc.)

## 🪂 Built-in templates

The following templates are built-in and available for use without any additional setup/files:

|   Template name   | Preview (light)                                | Preview (dark)                                |
|:-----------------:|------------------------------------------------|-----------------------------------------------|
|    `app-down`     | [![][app-down-light]][app-down-link]           | [![][app-down-dark]][app-down-link]           |
|      `cats`       | [![][cats-light]][cats-link]                   | [![][cats-dark]][cats-link]                   |
|   `connection`    | [![][connection-light]][connection-link]       | [![][connection-dark]][connection-link]       |
|      `ghost`      | [![][ghost-light]][ghost-link]                 | [![][ghost-dark]][ghost-link]                 |
| `hacker-terminal` | [![][hacker-terminal]][hacker-terminal-link]   | [![][hacker-terminal]][hacker-terminal-link]  |
|       `l7`        | [![][l7-light]][l7-link]                       | [![][l7-dark]][l7-link]                       |
|  `lost-in-space`  | [![][lost-in-space-light]][lost-in-space-link] | [![][lost-in-space-dark]][lost-in-space-link] |
|      `noise`      | [![][noise]][noise-link]                       | [![][noise]][noise-link]                      |
|     `orient`      | [![][orient-light]][orient-link]               | [![][orient-dark]][orient-link]               |
|     `shuffle`     | [![][shuffle-light]][shuffle-link]             | [![][shuffle-dark]][shuffle-link]             |
|      `win98`      | [![][win98-light]][win98-link]                 | [![][win98-dark]][win98-link]                 |

> [!NOTE]
> The `cats` template is the only one of those that fetches resources (the actual cat pictures) from external
> servers - all other templates are self-contained.

> [!TIP]
> If you need the **pre-rendered static error pages pack**, you can download it as a [zip][pages-pack-zip] or
> [tar.gz][pages-pack-tar-gz] archive.

[app-down-link]:https://tarampampam.github.io/error-pages/app-down/404.html
[app-down-light]:https://github.com/user-attachments/assets/135bac8c-983f-461c-97ba-e653e9b9adfe
[app-down-dark]:https://github.com/user-attachments/assets/c5f42b53-51cd-47c4-a22f-553d44d2a288
[cats-link]:https://tarampampam.github.io/error-pages/cats/404.html
[cats-light]:https://github.com/tarampampam/error-pages/assets/7326800/056cd00e-bc9a-4120-8325-310d7b0ebd1b
[cats-dark]:https://github.com/tarampampam/error-pages/assets/7326800/5689880b-f770-406c-81dd-2d28629e6f2e
[connection-link]:https://tarampampam.github.io/error-pages/connection/404.html
[connection-light]:https://github.com/tarampampam/error-pages/assets/7326800/099ecc2d-e724-4d9c-b5ed-66ddabd71139
[connection-dark]:https://github.com/tarampampam/error-pages/assets/7326800/3f03dc1b-c1ee-4a91-b3d7-e3b93c79020e
[ghost-link]:https://tarampampam.github.io/error-pages/ghost/404.html
[ghost-dark]:https://github.com/tarampampam/error-pages/assets/7326800/714482ab-f8c1-4455-8ae8-b2ae78f7a2c6
[ghost-light]:https://github.com/tarampampam/error-pages/assets/7326800/f253dfe7-96a0-4e96-915b-d4c544d4a237
[hacker-terminal-link]:https://tarampampam.github.io/error-pages/hacker-terminal/404.html
[hacker-terminal]:https://github.com/tarampampam/error-pages/assets/7326800/c197fc35-0844-43d0-9830-82440cee4559
[l7-link]:https://tarampampam.github.io/error-pages/l7/404.html
[l7-dark]:https://github.com/tarampampam/error-pages/assets/7326800/18e43ea3-6389-4459-be41-0fc6566a073f
[l7-light]:https://github.com/tarampampam/error-pages/assets/7326800/05f26669-94ec-40ce-8d67-a199cde54202
[lost-in-space-link]:https://tarampampam.github.io/error-pages/lost-in-space/404.html
[lost-in-space-dark]:https://github.com/tarampampam/error-pages/assets/7326800/debf87c0-6f27-41a8-b141-ee3464cbd6cc
[lost-in-space-light]:https://github.com/tarampampam/error-pages/assets/7326800/c347e63d-13a7-46d4-81b9-b25266819a1d
[noise-link]:https://tarampampam.github.io/error-pages/noise/404.html
[noise]:https://github.com/tarampampam/error-pages/assets/7326800/4cc5c3bd-6ebb-4e96-bee8-02d4ad4e7266
[orient-link]:https://tarampampam.github.io/error-pages/orient/404.html
[orient-dark]:https://github.com/tarampampam/error-pages/assets/7326800/bc2b0dad-c32c-4628-98f6-e3eab61dd1f2
[orient-light]:https://github.com/tarampampam/error-pages/assets/7326800/8fc0a7ea-694d-49ce-bb50-3ea032d52d1e
[shuffle-link]:https://tarampampam.github.io/error-pages/shuffle/404.html
[shuffle-dark]:https://github.com/tarampampam/error-pages/assets/7326800/7504b7c3-b0cb-4991-9ac2-759cd6c50fc0
[shuffle-light]:https://github.com/tarampampam/error-pages/assets/7326800/d2a73fc8-cf5f-4f42-bff8-cce33d8ae47e
[win98-link]:https://tarampampam.github.io/error-pages/win98/404.html
[win98-dark]:https://habrastorage.org/webt/bu/zt/5w/buzt5wsr-wixk0y8xjbxvepj0a8.png
[win98-light]:https://habrastorage.org/webt/pg/e8/f1/pge8f1ahyspmgu9vyh0jigvq_es.png

## 🚀 Installation

Download the latest binary for your OS/architecture from the [releases page][latest-release], or use the Docker image:

| Registry                          | Image                             |
|-----------------------------------|-----------------------------------|
| [GitHub Container Registry][ghcr] | `ghcr.io/tarampampam/error-pages` |
| [Quay.io][quay] (mirror)          | `quay.io/tarampampam/error-pages` |
| [Docker Hub][docker-hub] (mirror) | `tarampampam/error-pages`         |

> [!WARNING]
> Using the `latest` tag for Docker images is strongly discouraged, as it may introduce backward-incompatible changes
> during **major** upgrades. Use versioned tags in the `X`, `X.Y`, or `X.Y.Z` format instead.

> [!IMPORTANT]
> The app is distributed as two separate binaries - `error-pages` (HTTP server) and `builder`. Docker tags follow this
> convention:
> - `X.Y.Z` (and `X.Y`, `X`) - includes the HTTP server
> - `X.Y.Z-builder` (and `X.Y-builder`, `X-builder`) - includes the builder and a pre-rendered error pages pack

Supported image architectures - `linux/amd64`, `linux/arm/v7`, `linux/arm64`, `linux/ppc64le`, `linux/s390x`.
All images are signed with [Cosign][cosign] using keyless signing (GitHub OIDC).

### 📦 Helm chart

A Helm chart for Kubernetes is included with each release ([download][latest-helm-chart]), published on
[Artifact Hub][artifacthub], and also available via an OCI registry (Helm v3.8+ required):

```shell
helm install error-pages \
  oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version X.Y.Z
```

All supported chart values, examples, and usage instructions can be found at [Artifact Hub][artifacthub].

> Helm chart sources are located in the [deploy/helm](deploy/helm) directory of the repository.

[latest-release]:https://github.com/tarampampam/error-pages/releases/latest
[ghcr]:https://github.com/tarampampam/error-pages/pkgs/container/error-pages
[docker-hub]:https://hub.docker.com/r/tarampampam/error-pages
[quay]:https://quay.io/repository/tarampampam/error-pages?tab=tags
[pages-pack-zip]:https://github.com/tarampampam/error-pages/releases/latest/download/error-pages-static.zip
[pages-pack-tar-gz]:https://github.com/tarampampam/error-pages/releases/latest/download/error-pages-static.tar.gz
[cosign]:https://github.com/sigstore/cosign
[latest-helm-chart]:https://github.com/tarampampam/error-pages/releases/latest/download/helm-chart.tgz
[artifacthub]:https://artifacthub.io/packages/helm/error-pages/error-pages

## 🛠 Integration guides

- **Standalone** server or builder usage:
  - [🚀 Start the HTTP server with a custom template (theme)](docs/guides/theming.md)
  - [🚀 Generate error pages using built-in or custom templates](docs/guides/builder.md)

- With **Nginx**:
  - [🚀 Customize error pages in your own Nginx Docker image](docs/guides/nginx_image.md)
  - [🚀 Use Nginx as a reverse proxy with custom error pages](docs/guides/nginx_upstream.md)

- With **Caddy**:
  - [🚀 Customize error pages in your own Caddy Docker image](docs/guides/caddy_image.md)
  - [🚀 Use Caddy as a reverse proxy with custom error pages](docs/guides/caddy_upstream.md)

- With **Traefik**:
  - [🚀 Use Traefik with a local Docker Compose setup](docs/guides/traefik_docker.md)

- With **Kubernetes**:
  - [🚀 Use error-pages as the **ingress-nginx** default backend](docs/guides/k8s_ingress_nginx.md)
  - [🚀 Use error-pages with **Traefik** in Kubernetes](docs/guides/k8s_ingress_traefik.md)
  - [🚀 Use error-pages with **NGINX Gateway Fabric** in Kubernetes](docs/guides/k8s_ingress_ngf.md)
  - [🚀 Use error-pages with **Envoy Gateway** in Kubernetes](docs/guides/k8s_ingress_envoy.md)
  - [🚀 Use error pages with **HAProxy Ingress** in Kubernetes](docs/guides/k8s_ingress_haproxy.md)

## 🦾 Performance

Measured on loopback (`127.0.0.1`), single connection, no artificial load
(`wrk -t1 -c1 -d5s --latency http://127.0.0.1:8080/...`, less is better):

|      Format | p50 (typical response time) | p90 (90% of responses complete within this time) |
|------------:|:---------------------------:|:------------------------------------------------:|
|        HTML |         **121 µs**          |                      262 µs                      |
|        JSON |          **51 µs**          |                      75 µs                       |
|         XML |          **48 µs**          |                      73 µs                       |
|  Plain text |          **47 µs**          |                      68 µs                       |
| HTML + gzip |         **2.4 ms**          |                      3.1 ms                      |
| JSON + gzip |         **256 µs**          |                      510 µs                      |

> [!NOTE]
> HTML responses are large (full rendered template, ~65 KB), which is why gzip compression takes noticeably more
> time there. JSON/XML/text are compact structured responses, so they are fastest overall.

## 💻 Command-line usage

For detailed instructions on using the HTTP server and the static site generator, including all supported environment
variables and usage examples, check the [CLI documentation](docs/CLI.md).

## 🔍 How the server handles requests

The three most important things to understand about how the server behaves - how it determines which error page to
show, which format to return, and what request context it can expose.

### How the error code is resolved

The server picks the HTTP status code from the **first** matching source:

1. Path: `/404`, `/404.html`, `/404.json`, `/503.xml`, etc.
2. `X-Code` request header
3. Default: `--default-error-page` (or `DEFAULT_ERROR_PAGE`, default: 404)

### How the response format is determined

The response format is picked from the **first** matching source:

1. Path extension: `.html`, `.htm`, `.json`, `.xml`, `.txt`
2. `Content-Type` request header
3. `X-Format` request header (e.g. `X-Format: application/json`)
4. `Accept` request header
5. Default: **plain text**

Supported formats: `HTML`, `JSON`, `XML`, `plain text`.

### Service endpoints

The following HTTP endpoints can be used for health checks, monitoring, or other purposes:

| Path                                           | Description                              |
|------------------------------------------------|------------------------------------------|
| `/healthz`, `/health`, `/health/live`, `/live` | Liveness probe - always returns `200 OK` |
| `/version`                                     | Returns `{"version":"..."}` as JSON      |

### Response headers

Every error page response includes the following headers automatically:

| Header             | Value                                     | Notes                                              |
|--------------------|-------------------------------------------|----------------------------------------------------|
| `Content-Type`     | e.g. `text/html; charset=utf-8`           | Format-dependent                                   |
| `Content-Length`   | Response body size in bytes               | Always set                                         |
| `X-Robots-Tag`     | `noindex, nofollow, nosnippet, noarchive` | Prevents error pages from being indexed            |
| `Retry-After`      | `120`                                     | Only for limited set of status codes               |
| `Content-Encoding` | `gzip`                                    | Only when the client sends `Accept-Encoding: gzip` |

Headers listed in `--proxy-headers` (default: `X-Request-Id`, `X-Trace-Id`, `X-Correlation-Id`,
`X-Amzn-Trace-Id`) are copied from the incoming request to the response when present.

### HTTP status code of the response

By default, the server always responds with **HTTP 200**, even when rendering error pages. This is the correct
behavior when a reverse proxy (Nginx, Traefik, ingress-nginx) intercepts upstream error responses and replaces
only the body - the proxy itself sends the original error status code back to the client.

When error-pages is used as a **direct backend** - e.g. a catch-all route, a Kubernetes default backend, or
standalone testing - it must return the correct status code itself. Enable `--send-same-http-code`
(or env `SEND_SAME_HTTP_CODE=true`) to make the HTTP response status match the error code being rendered.

## 📝 Templating and Localization

For detailed instructions on using custom templates and localization features, see the
[templating documentation](docs/templating.md).

## 🔧 Development

### Requirements

- [**Go 1.26+**](https://go.dev/doc/install) for building from source and running tests
- Optional: [golangci-lint](https://golangci-lint.run/docs/welcome/install/local/) for linting
- Optional: [docker](https://docs.docker.com/engine/install/debian/#install-using-the-convenience-script) for testing
  the Docker image locally
- Optional: [helm](https://helm.sh/docs/intro/install/) + [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) for
  testing the Helm chart locally in Kubernetes
- Optional: [helm-docs](https://github.com/norwoodj/helm-docs/releases/latest) for generating Helm chart documentation
- Optional: [watchexec](https://github.com/watchexec/watchexec/releases/latest) for live reloading the server during
  development

**Commands**:

```shell
go generate -skip readme ./...            # (re)generate code, except docs
go generate ./...                         # (re)generate everything
go build ./cmd/error-pages/ && go build ./cmd/builder/ # build both binaries
go test -race ./...                       # run all tests
golangci-lint run --fix                   # lint the code and apply any available auto fixes
helm-docs -c ./deploy/helm/ -t README.tpl.md -o README.md # regenerate chart readme file

# run a live reloading server (useful for testing template changes)
watchexec -r -- go run ./cmd/error-pages/ --show-details

# run before every vibe-coding session
your_ai_tool --prompt "explain why AI Coding Agents are doing shit by default"
```

## 🧑‍🤝‍🧑 Contributors

I want to say a big thank you to everyone who contributed to this project:

[![contributors](https://contrib.rocks/image?repo=tarampampam/error-pages)][contributors]

[contributors]:https://github.com/tarampampam/error-pages/graphs/contributors

## 🤝 Contributing & AI-assisted development

Missing a feature? Found a bug you want fixed? Pull requests are welcome - and yes, you are explicitly invited to
try implementing it with an AI coding agent.

To give the agent a fighting chance at producing something that fits this codebase, the repo ships an
[`AGENTS.md`](AGENTS.md) - a structured reference covering project layout, build commands, code style, generated
files, hard prohibitions, and the full post-change workflow. It is written **for the agent**. Most modern agents
pick it up automatically.

**Review every single changed line yourself**. Understand it. Be able to defend it in code review. If you cannot
explain why a line is there and why it is correct, do not open the PR. _"The agent wrote it"_ is not an answer.
The author of a PR is the human who opens it, not the model (at least, I hope so).

I write my own code by hand and encourage you to do the same when you can. AI is a tool, not an excuse to skip the
thinking. Trust, but verify - and verify hard.

### 🤖 Setup for AI agents

This repository follows the [agents.md](https://agents.md/) open standard. The canonical instructions live in
[`AGENTS.md`](AGENTS.md).

## 👾 Support

[![Issues][badge-issues]][issues]
[![Issues][badge-prs]][prs]

If you encounter any bugs in the project, please [create an issue][new-issue] in this repository.

[badge-issues]:https://img.shields.io/github/issues/tarampampam/error-pages.svg?maxAge=45
[badge-prs]:https://img.shields.io/github/issues-pr/tarampampam/error-pages.svg?maxAge=45
[issues]:https://github.com/tarampampam/error-pages/issues
[prs]:https://github.com/tarampampam/error-pages/pulls
[new-issue]:https://github.com/tarampampam/error-pages/issues/new/choose

## 📖 License

This is open-sourced software licensed under the [MIT License][license].

[license]:https://github.com/tarampampam/error-pages/blob/master/LICENSE
