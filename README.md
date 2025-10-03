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

One day, you might want to replace the standard error pages of your HTTP server or K8S cluster with something more
original and attractive. That's why this repository was created :) It contains:

- A simple error page generator written in Go
- Single-page error templates (themes) with various designs (located in the [templates][templates-dir] directory) that
  you can customize as you wish
- A fast and lightweight HTTP server is available as a single binary file and Docker image. It includes built-in error
  page templates from this repository. You don't need anything except the compiled binary file or Docker image
- Pre-generated error pages (sources can be [found here][preview-sources], and the [**demo** is always
  accessible here][preview-demo])

[preview-sources]:https://github.com/tarampampam/error-pages/tree/gh-pages
[preview-demo]:https://tarampampam.github.io/error-pages/
[templates-dir]:https://github.com/tarampampam/error-pages/tree/master/templates

## 🔥 Features List

- HTTP server written in Go, utilizing the extremely fast [FastHTTP][fasthttp] and in-memory caching
  - Respects the `Content-Type` HTTP header (and `X-Format`) value, responding with the corresponding format
    (supported formats: `json`, `xml`, and `plaintext`)
  - Error pages are configured to be excluded from search engine indexing (using meta tags and HTTP headers) to
    prevent SEO issues on your website
  - HTML content (including CSS, SVG, and JS) is minified on the fly
  - Logs written in `json` format
  - Contains a health check endpoint (`/healthz`)
  - Consumes very few resources and is suitable for use in resource-constrained environments
- Lightweight Docker image, distroless, and uses an unprivileged user by default
- [Go-template](https://pkg.go.dev/text/template) tags are allowed in the templates
- Ready for integration with [Traefik][traefik], [Ingress-nginx][ingress-nginx], and more
- Error pages can be embedded into your own Docker image with `nginx` in a few simple steps
- Fully configurable
- Distributed as a Docker image and compiled binary files
- Localized HTML error pages (🇺🇸, 🇫🇷, 🇺🇦, 🇷🇺, 🇵🇹, 🇳🇱, 🇩🇪, 🇪🇸, 🇨🇳, 🇮🇩, 🇵🇱, 🇰🇷, 🇭🇺, 🇳🇴, 🇷🇴) - translation process
  [described here][l10n-dir] - other translations are welcome!

[fasthttp]:https://github.com/valyala/fasthttp
[traefik]:https://github.com/traefik/traefik
[l10n-dir]:https://github.com/tarampampam/error-pages/tree/master/l10n

## 🧩 Install

Download the latest binary file for your OS/architecture from the [releases page][latest-release] or use our Docker image:

| Registry                          | Image                             |
|-----------------------------------|-----------------------------------|
| [GitHub Container Registry][ghcr] | `ghcr.io/tarampampam/error-pages` |
| [Docker Hub][docker-hub] (mirror) | `tarampampam/error-pages`         |

> [!IMPORTANT]
> Using the `latest` tag for the Docker image is highly discouraged due to potential backward-incompatible changes
> during **major** upgrades. Please use tags in the `X.Y.Z` format.

💣 **Or** you can also download the **already rendered** error pages pack as a [zip][pages-pack-zip] or
[tar.gz][pages-pack-tar-gz] archive.

[latest-release]:https://github.com/tarampampam/error-pages/releases/latest
[docker-hub]:https://hub.docker.com/r/tarampampam/error-pages
[ghcr]:https://github.com/tarampampam/error-pages/pkgs/container/error-pages
[pages-pack-zip]:https://github.com/tarampampam/error-pages/zipball/gh-pages/
[pages-pack-tar-gz]:https://github.com/tarampampam/error-pages/tarball/gh-pages/

## 🪂 Templates (themes)

The following templates are built-in and available for use without any additional setup:

| Template name     | Preview (light)                                | Preview (dark)                                |
|-------------------|------------------------------------------------|-----------------------------------------------|
| `app-down`        | [![][app-down-light]][app-down-link]           | [![][app-down-dark]][app-down-link]           |
| `cats`            | [![][cats-light]][cats-link]                   | [![][cats-dark]][cats-link]                   |
| `connection`      | [![][connection-light]][connection-link]       | [![][connection-dark]][connection-link]       |
| `ghost`           | [![][ghost-light]][ghost-link]                 | [![][ghost-dark]][ghost-link]                 |
| `hacker-terminal` | [![][hacker-terminal]][hacker-terminal-link]   | [![][hacker-terminal]][hacker-terminal-link]  |
| `l7`              | [![][l7-light]][l7-link]                       | [![][l7-dark]][l7-link]                       |
| `lost-in-space`   | [![][lost-in-space-light]][lost-in-space-link] | [![][lost-in-space-dark]][lost-in-space-link] |
| `noise`           | [![][noise]][noise-link]                       | [![][noise]][noise-link]                      |
| `orient`          | [![][orient-light]][orient-link]               | [![][orient-dark]][orient-link]               |
| `shuffle`         | [![][shuffle-light]][shuffle-link]             | [![][shuffle-dark]][shuffle-link]             |
| `win98`           | [![][win98-light]][win98-link]                 | [![][win98-dark]][win98-link]                 |

> [!NOTE]
> The `cats` template is the only one of those that fetches resources (the actual cat pictures) from external
> servers - all other templates are self-contained.

[app-down-link]:https://tarampampam.github.io/error-pages/app-down/404.html
[app-down-light]:https://github.com/tarampampam/error-pages/assets/7326800/ad4b4fd7-7c7b-4bdc-a6b6-44f9ba7f77ca
[app-down-dark]:https://github.com/tarampampam/error-pages/assets/7326800/4e668a56-a4c4-47cd-ac4d-b6b45db54ab8
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

## 🛠 Usage scenarios

### HTTP server starting, utilizing either a binary file or Docker image

First, ensure you have a precompiled binary file on your machine or have Docker/Podman installed. Next, start the
server with the following command:

```bash
$ ./error-pages serve
# --- or ---
$ docker run --rm -p '8080:8080/tcp' tarampampam/error-pages serve
```

That's it! The server will begin running and listen on address `0.0.0.0` and port `8080`. Access error pages using
URLs like `http://127.0.0.1:8080/{page_code}.html`.

To retrieve different error page codes using a static URL, use the `X-Code` HTTP header:

```bash
$ curl -H 'X-Code: 500' http://127.0.0.1:8080/
```

The server respects the `Content-Type` HTTP header (and `X-Format`), delivering responses in requested formats
such as HTML, XML, JSON, and PlainText. Customization of these formats is possible via CLI flags or environment
variables.

For integration with [ingress-nginx][ingress-nginx] or debugging purposes, start the server with `--show-details`
(or set the environment variable `SHOW_DETAILS=true`) to enrich error pages (including JSON and XML responses)
with upstream proxy information.

Switch themes using the `TEMPLATE_NAME` environment variable or the `--template-name` flag; available templates
are detailed in the readme file below.

> [!TIP]
> Use the `--rotation-mode` flag or the `TEMPLATES_ROTATION_MODE` environment variable to automate theme
> rotation. Available modes include `random-on-startup`, `random-on-each-request`, `random-hourly`,
> and `random-daily`.

To proxy HTTP headers from requests to responses, utilize the `--proxy-headers` flag or environment variable
(comma-separated list of headers).

### 🔌 Integrations with Traefik, Nginx, Kubernetes (and more)

<details>
  <summary><strong>🚀 Start the HTTP server with my custom template (theme)</strong></summary>

First, create your own template file, for example `my-super-theme.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <title>{{ code }}</title>
</head>
<body>
  <h1>YEAH! {{ message }}: {{ description }}</h1>
</body>
</html>
```

And simply start the server with the following command:

```bash
$ docker run --rm \
  -v "$(pwd)/my-super-theme.html:/opt/my-template.html:ro" \
  -p '8080:8080/tcp' ghcr.io/tarampampam/error-pages:3 serve \
    --add-template /opt/my-template.html \
    --template-name my-template
# --- or ---
$ ./error-pages serve \
  --add-template /opt/my-template.html \
  --template-name my-template
```

And test it:

```bash
$ curl -H "Accept: text/html" http://127.0.0.1:8080/503

<!DOCTYPE html>
<html lang="en">
<head>
  <title>503</title>
</head>
<body>
  <h1>YEAH! Service Unavailable: The server is temporarily overloading or down</h1>
</body>
</html>
```

</details>

<details>
  <summary><strong>🚀 Generate a set of error pages using built-in or my own template</strong></summary>

Generating a set of error pages is straightforward. If you prefer to use your own template, start by crafting it.
Create a file like this:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <title>{{ code }}</title>
</head>
<body>
  <h1>{{ message }}: {{ description }}</h1>
</body>
</html>
```

Save it as `my-template.html` and use it as your custom template. Then, generate your error pages using the command:

```bash
$ mkdir -p /path/to/output
$ ./error-pages build --add-template /path/to/your/my-template.html --target-dir /path/to/output
```

This will create error pages based on your template in the specified output directory:

```bash
$ cd /path/to/output && tree .
├── my-template
│   ├── 400.html
│   ├── 401.html
│   ├── 403.html
│   ├── 404.html
│   ├── 405.html
│   ├── 407.html
│   ├── 408.html
│   ├── 409.html
│   ├── 410.html
│   ├── 411.html
│   ├── 412.html
│   ├── 413.html
│   ├── 416.html
│   ├── 418.html
│   ├── 429.html
│   ├── 500.html
│   ├── 502.html
│   ├── 503.html
│   ├── 504.html
│   └── 505.html
…

$ cat my-template/403.html
<!DOCTYPE html>
<html lang="en">
<head>
  <title>403</title>
</head>
<body>
  <h1>Forbidden: Access is forbidden to the requested page</h1>
</body>
</html>
```

</details>

<details>
  <summary><strong>🚀 Customize error pages within your own Nginx Docker image</strong></summary>

To create this cocktail, we need two components:

- Nginx configuration file
- A Dockerfile to build the image

Let's start with the Nginx configuration file:

```nginx
# File: nginx.conf

server {
  listen      80;
  server_name localhost;

  error_page 401 /_error-pages/401.html;
  error_page 403 /_error-pages/403.html;
  error_page 404 /_error-pages/404.html;
  error_page 500 /_error-pages/500.html;
  error_page 502 /_error-pages/502.html;
  error_page 503 /_error-pages/503.html;

  location ^~ /_error-pages/ {
    internal;
    root /usr/share/nginx/errorpages;
  }

  location / {
    root  /usr/share/nginx/html;
    index index.html index.htm;
  }
}
```

And the Dockerfile:

```dockerfile
FROM docker.io/library/nginx:1.27-alpine

# override default Nginx configuration
COPY --chown=nginx ./nginx.conf /etc/nginx/conf.d/default.conf

# copy statically built error pages from the error-pages image
# (instead of `ghost` you may use any other template)
COPY --chown=nginx \
     --from=ghcr.io/tarampampam/error-pages:3 \
     /opt/html/ghost /usr/share/nginx/errorpages/_error-pages
```

Now, we can build the image:

```bash
$ docker build --tag your-nginx:local -f ./Dockerfile .
```

And voilà! Let's start the image and test if everything is working as expected:

```bash
$ docker run --rm -p '8081:80/tcp' your-nginx:local

$ curl http://127.0.0.1:8081/foobar | head -n 15 # in another terminal
```

</details>

<details>
  <summary><strong>🚀 Usage with Traefik and local Docker Compose</strong></summary>

Instead of thousands of words, let's take a look at one compose file:

```yaml
# file: compose.yml (or docker-compose.yml)

services:
  traefik:
    image: docker.io/library/traefik:v3.1
    command:
      #- --log.level=DEBUG
      - --api.dashboard=true # activate dashboard
      - --api.insecure=true # enable the API in insecure mode
      - --providers.docker=true # enable Docker backend with default settings
      - --providers.docker.exposedbydefault=false # do not expose containers by default
      - --entrypoints.web.address=:80 # --entrypoints.<name>.address for ports, 80 (i.e., name = web)
    ports:
      - "80:80/tcp" # HTTP (web)
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    labels:
      traefik.enable: true
      # dashboard
      traefik.http.routers.traefik.rule: Host(`traefik.localtest.me`)
      traefik.http.routers.traefik.service: api@internal
      traefik.http.routers.traefik.entrypoints: web
      traefik.http.routers.traefik.middlewares: error-pages-middleware
    depends_on:
      error-pages: {condition: service_healthy}

  error-pages:
    image: ghcr.io/tarampampam/error-pages:3 # using the latest tag is highly discouraged
    environment:
      TEMPLATE_NAME: l7 # set the error pages template
    labels:
      traefik.enable: true
      # use as "fallback" for any NON-registered services (with priority below normal)
      traefik.http.routers.error-pages-router.rule: HostRegexp(`.+`)
      traefik.http.routers.error-pages-router.priority: 10
      # should say that all of your services work on https
      traefik.http.routers.error-pages-router.entrypoints: web
      traefik.http.routers.error-pages-router.middlewares: error-pages-middleware
      # "errors" middleware settings
      traefik.http.middlewares.error-pages-middleware.errors.status: 400-599
      traefik.http.middlewares.error-pages-middleware.errors.service: error-pages-service
      traefik.http.middlewares.error-pages-middleware.errors.query: /{status}.html
      # define service properties
      traefik.http.services.error-pages-service.loadbalancer.server.port: 8080

  nginx-or-any-another-service:
    image: docker.io/library/nginx:1.27-alpine
    labels:
      traefik.enable: true
      traefik.http.routers.test-service.rule: Host(`test.localtest.me`)
      traefik.http.routers.test-service.entrypoints: web
      traefik.http.routers.test-service.middlewares: error-pages-middleware
```

After executing `docker compose up` in the same directory as the `compose.yml` file, you can:

- Open the Traefik dashboard [at `traefik.localtest.me`](http://traefik.localtest.me/dashboard/#/)
- [View customized error pages on the Traefik dashboard](http://traefik.localtest.me/foobar404)
- Open the nginx index page [at `test.localtest.me`](http://test.localtest.me/)
- View customized error pages for non-existent [pages](http://test.localtest.me/404) and [domains](http://404.localtest.me/)

Isn't this kind of magic? 😀

</details>

<details>
  <summary><strong>🚀 Kubernetes (K8s) & Ingress Nginx</strong></summary>

Error-pages can be configured to work with the [ingress-nginx][ingress-nginx] helm chart in Kubernetes.

- Set the `custom-http-errors` config value
- Enable default backend
- Set the default backend image

```yaml
controller:
  config:
    custom-http-errors: >-
      401,403,404,500,501,502,503

defaultBackend:
  enabled: true
  image:
    repository: ghcr.io/tarampampam/error-pages
    tag: '3' # using the latest tag is highly discouraged
  extraEnvs:
  - name: TEMPLATE_NAME # Optional: change the default theme
    value: l7
  - name: SHOW_DETAILS # Optional: enables the output of additional information on error pages
    value: 'true'
```

</details>

<details>
  <summary><strong>🚀 Kubernetes (K8s) & Ingress Traefik</strong></summary>

There are various ways to set up "error pages" in Kubernetes with Traefik. One of the most common scenarios is when
you already have Traefik installed as an Ingress Controller with all the necessary CRDs. In this case, you still
need to install the "error pages" as a separate service, register a middleware that will use it, and apply this
middleware to all relevant routers.

> To install Traefik using Helm, you may add the following lines to your `Chart.yaml` file:
>
> ```yaml
> dependencies:
> - name: traefik
>   version: 34.1.0 # change to the latest version
>   repository: https://helm.traefik.io/traefik
> ```

I prefer to install each component in a separate namespace and use Helm to manage the installation process. So
before we begin, let's define the following settings in the `values.yaml` file:

```yaml
errorPages:
  enabled: true
  appName: error-pages
  namespace: error-pages
  version: 3.3.1 # https://github.com/tarampampam/error-pages/releases
  themeName: shuffle
```

Next, create the following Helm chart templates:

```yaml
# file: error-pages/namespace.yaml

{{ with .Values.errorPages }}
{{- if .enabled }}
apiVersion: v1
kind: Namespace

metadata: {name: "{{ .namespace }}"}
{{- end }}
{{- end }}
```

```yaml
# file: error-pages/deployment.yaml

{{ with .Values.errorPages }}
{{- if .enabled }}
apiVersion: apps/v1
kind: Deployment

metadata:
  name: "{{ .appName }}"
  namespace: {{ .namespace }}
  labels: {app: "{{ .appName }}"}

spec:
  replicas: 1
  selector: {matchLabels: {app: "{{ .appName }}"}}
  template:
    metadata: {labels: {app: "{{ .appName }}"}}
    spec:
      automountServiceAccountToken: false
      containers:
        - name: "{{ .appName }}"
          image: "ghcr.io/tarampampam/error-pages:{{ .version | default "latest" }}"
          env:
            - {name: TEMPLATE_NAME, value: "{{ .themeName | default "app-down" }}"}
          securityContext:
            runAsNonRoot: true
            runAsUser: 10001
            runAsGroup: 10001
            readOnlyRootFilesystem: true
          ports:
            - {name: http, containerPort: 8080, protocol: TCP}
          livenessProbe:
            httpGet: {port: http, path: /healthz}
            periodSeconds: 10
          readinessProbe:
            httpGet: {port: http, path: /healthz}
            periodSeconds: 10
          resources:
            limits: {memory: 64Mi, cpu: 200m} # change if needed
            requests: {memory: 16Mi, cpu: 20m}
{{- end }}
{{- end }}
```

```yaml
# file: error-pages/service.yaml

{{ with .Values.errorPages }}
{{- if .enabled }}
apiVersion: v1
kind: Service

metadata:
  name: {{ .appName }}-service
  namespace: {{ .namespace }}
  labels: {app: "{{ .appName }}"}

spec:
  type: ClusterIP
  selector: {app: "{{ .appName }}"}
  ports: [{name: http, protocol: TCP, port: 8080, targetPort: 8080}]
{{- end }}
{{- end }}
```

```yaml
# file: error-pages/middleware.yaml

{{ with .Values.errorPages }}
{{- if .enabled }}
apiVersion: traefik.io/v1alpha1
kind: Middleware

metadata:
  name: {{ .appName }}
  namespace: {{ .namespace }}

spec: # https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-middleware
  errors:
    status: ["401", "403", "404", "500-599"]
    service: {name: "{{ .appName }}-service", port: 8080}
    query: "/{status}.html"
{{- end }}
{{- end }}
```

If everything was configured correctly, you should see the new middleware in your Traefik dashboard after applying
updated Helm chart:

![traefik-dashboard-middleware](https://habrastorage.org/webt/dj/to/qy/djtoqy20fpi2_qqabzeyltwlenw.png)

Since our middleware is in a separate namespace, and in Traefik >=2.5 cross-namespace references for resources
like middlewares are restricted by default, we need to enable this feature. To do so, add the following lines to
your Traefik Helm chart values:

```diff
 traefik:
   # ...
-  globalArguments: []
+  globalArguments: ["--providers.kubernetescrd.allowCrossNamespace=true"]
   # ...
```

Or if you use the [Traefik Helm chart](https://helm.traefik.io/traefik) >= 36.*:

```diff
 traefik:
   # ...
+  providers:
+    kubernetesCRD:
+      allowCrossNamespace: true
   # ...
```

Now, you can apply the middleware to the necessary ingress routes:

```diff
 apiVersion: traefik.io/v1alpha1
 kind: IngressRoute

 metadata:
   name: some-app-http
   namespace: some-app

 spec: # https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-ingressroute
   entryPoints: [websecure]
   routes:
     - match: Host(`my.awesome-site.com`) && PathPrefix(`/`)
       services: [{name: "some-app-service", namespace: some-app, port: 8080}]
+      {{- with $.Values.errorPages }}{{ if .enabled }}
+      middlewares: [{name: "{{ .appName }}", namespace: "{{ .namespace }}"}]
+      {{- end }}{{ end }}
```

Although this approach is quite verbose, it allows for full control over the configuration. If you have a
better alternative, feel free to submit a PR!

</details>

## 🦾 Performance

Hardware used:

- 12th Gen Intel® Core™ i7-1260P (16 cores)
- 32 GiB RAM

RPS: **~180k** 🔥 requests served without any errors, with peak memory usage ~60 MiB under the default configuration

<details>
  <summary>Performance test details (click to expand)</summary>

```shell
$ ulimit -aH | grep file
core file size              (blocks, -c) unlimited
file size                   (blocks, -f) unlimited
open files                          (-n) 1048576
file locks                          (-x) unlimited

$ go build ./cmd/error-pages/ && ./error-pages --log-level warn serve

$ ./error-pages perftest # in separate terminal
Starting the test to bomb ONE PAGE (code). Please, be patient...
Test completed successfully. Here is the output:

Running 15s test @ http://127.0.0.1:8080/
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.52ms    6.43ms  94.34ms   85.44%
    Req/Sec    15.76k     2.83k   29.64k    69.20%
  2839632 requests in 15.09s, 32.90GB read
Requests/sec: 188185.61
Transfer/sec:      2.18GB

Starting the test to bomb DIFFERENT PAGES (codes). Please, be patient...
Test completed successfully. Here is the output:

Running 15s test @ http://127.0.0.1:8080/
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.75ms   13.71ms 252.66ms   91.94%
    Req/Sec    14.06k     3.25k   26.39k    71.98%
  2534473 requests in 15.10s, 29.22GB read
Requests/sec: 167899.78
Transfer/sec:      1.94GB
```

</details>

<!--GENERATED:CLI_DOCS-->
<!-- Documentation inside this block generated by github.com/urfave/cli-docs/v3; DO NOT EDIT -->
## CLI interface

Usage:

```bash
$ error-pages [GLOBAL FLAGS] [COMMAND] [COMMAND FLAGS] [ARGUMENTS...]
```

Global flags:

| Name               | Description                           | Default value | Environment variables |
|--------------------|---------------------------------------|:-------------:|:---------------------:|
| `--log-level="…"`  | Logging level (debug/info/warn/error) |    `info`     |      `LOG_LEVEL`      |
| `--log-format="…"` | Logging format (console/json)         |   `console`   |     `LOG_FORMAT`      |

### `serve` command (aliases: `s`, `server`, `http`)

Please start the HTTP server to serve the error pages. You can configure various options - please RTFM :D.

Usage:

```bash
$ error-pages [GLOBAL FLAGS] serve [COMMAND FLAGS] [ARGUMENTS...]
```

The following flags are supported:

| Name                                                  | Description                                                                                                                                                                                                                                                                                                               |               Default value               |    Environment variables    |
|-------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-----------------------------------------:|:---------------------------:|
| `--listen="…"` (`-l`)                                 | The HTTP server will listen on this IP (v4 or v6) address (set 127.0.0.1/::1 for localhost, 0.0.0.0 to listen on all interfaces, or specify a custom IP)                                                                                                                                                                  |                 `0.0.0.0`                 |        `LISTEN_ADDR`        |
| `--port="…"` (`-p`)                                   | The TCP port number for the HTTP server to listen on (0-65535)                                                                                                                                                                                                                                                            |                  `8080`                   |        `LISTEN_PORT`        |
| `--add-template="…"`                                  | To add a new template, provide the path to the file using this flag (the filename without the extension will be used as the template name)                                                                                                                                                                                |                   `[]`                    |       `ADD_TEMPLATE`        |
| `--disable-template="…"`                              | Disable the specified template by its name (useful to disable the built-in templates and use only custom ones)                                                                                                                                                                                                            |                   `[]`                    |           *none*            |
| `--add-code="…"`                                      | To add a new HTTP status code, provide the code and its message/description using this flag (the format should be '%code%=%message%/%description%'; the code may contain a wildcard '*' to cover multiple codes at once, for example, '4**' will cover all 4xx codes unless a more specific code is described previously) |                  `map[]`                  |           *none*            |
| `--json-format="…"`                                   | Override the default error page response in JSON format (Go templates are supported; the error page will use this template if the client requests JSON content type)                                                                                                                                                      |                                           |   `RESPONSE_JSON_FORMAT`    |
| `--xml-format="…"`                                    | Override the default error page response in XML format (Go templates are supported; the error page will use this template if the client requests XML content type)                                                                                                                                                        |                                           |    `RESPONSE_XML_FORMAT`    |
| `--plaintext-format="…"`                              | Override the default error page response in plain text format (Go templates are supported; the error page will use this template if the client requests plain text content type or does not specify any)                                                                                                                  |                                           | `RESPONSE_PLAINTEXT_FORMAT` |
| `--template-name="…"` (`-t`, `--template`, `--theme`) | Name of the template to use for rendering error pages (built-in templates: app-down, cats, connection, ghost, hacker-terminal, l7, lost-in-space, noise, orient, shuffle, win98)                                                                                                                                          |                `app-down`                 |       `TEMPLATE_NAME`       |
| `--disable-l10n`                                      | Disable localization of error pages (if the template supports localization)                                                                                                                                                                                                                                               |                  `false`                  |       `DISABLE_L10N`        |
| `--default-error-page="…"`                            | The code of the default (index page, when a code is not specified) error page to render                                                                                                                                                                                                                                   |                   `404`                   |    `DEFAULT_ERROR_PAGE`     |
| `--send-same-http-code`                               | The HTTP response should have the same status code as the requested error page (by default, every response with an error page will have a status code of 200)                                                                                                                                                             |                  `false`                  |    `SEND_SAME_HTTP_CODE`    |
| `--show-details`                                      | Show request details in the error page response (if supported by the template)                                                                                                                                                                                                                                            |                  `false`                  |       `SHOW_DETAILS`        |
| `--proxy-headers="…"`                                 | HTTP headers listed here will be proxied from the original request to the error page response (comma-separated list)                                                                                                                                                                                                      | `X-Request-Id,X-Trace-Id,X-Amzn-Trace-Id` |    `PROXY_HTTP_HEADERS`     |
| `--rotation-mode="…"`                                 | Templates automatic rotation mode (disabled/random-on-startup/random-on-each-request/random-hourly/random-daily)                                                                                                                                                                                                          |                `disabled`                 |  `TEMPLATES_ROTATION_MODE`  |
| `--read-buffer-size="…"`                              | Per-connection buffer size in bytes for reading requests, this also limits the maximum header size (increase this buffer if your clients send multi-KB Request URIs and/or multi-KB headers (e.g., large cookies), note that increasing this value will increase memory consumption)                                      |                  `5120`                   |     `READ_BUFFER_SIZE`      |
| `--disable-minification`                              | Disable the minification of HTML pages, including CSS, SVG, and JS (may be useful for debugging)                                                                                                                                                                                                                          |                  `false`                  |   `DISABLE_MINIFICATION`    |

### `build` command (aliases: `b`)

Build the static error pages and put them into a specified directory.

Usage:

```bash
$ error-pages [GLOBAL FLAGS] build [COMMAND FLAGS] [ARGUMENTS...]
```

The following flags are supported:

| Name                                        | Description                                                                                                                                                                                                                                                                                                               | Default value |  Environment variables |
|---------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------:|:----------------------:|
| `--add-template="…"`                        | To add a new template, provide the path to the file using this flag (the filename without the extension will be used as the template name)                                                                                                                                                                                |     `[]`      |     `ADD_TEMPLATE`     |
| `--disable-template="…"`                    | Disable the specified template by its name (useful to disable the built-in templates and use only custom ones)                                                                                                                                                                                                            |     `[]`      |         *none*         |
| `--add-code="…"`                            | To add a new HTTP status code, provide the code and its message/description using this flag (the format should be '%code%=%message%/%description%'; the code may contain a wildcard '*' to cover multiple codes at once, for example, '4**' will cover all 4xx codes unless a more specific code is described previously) |    `map[]`    |         *none*         |
| `--disable-l10n`                            | Disable localization of error pages (if the template supports localization)                                                                                                                                                                                                                                               |    `false`    |     `DISABLE_L10N`     |
| `--index` (`-i`)                            | Generate index.html file with links to all error pages                                                                                                                                                                                                                                                                    |    `false`    |         *none*         |
| `--target-dir="…"` (`--out`, `--dir`, `-o`) | Directory to put the built error pages into                                                                                                                                                                                                                                                                               |      `.`      |         *none*         |
| `--disable-minification`                    | Disable the minification of HTML pages, including CSS, SVG, and JS (may be useful for debugging)                                                                                                                                                                                                                          |    `false`    | `DISABLE_MINIFICATION` |

### `healthcheck` command (aliases: `chk`, `health`, `check`)

Health checker for the HTTP server. The use case - docker health check.

Usage:

```bash
$ error-pages [GLOBAL FLAGS] healthcheck [COMMAND FLAGS] [ARGUMENTS...]
```

The following flags are supported:

| Name                | Description                                   | Default value | Environment variables |
|---------------------|-----------------------------------------------|:-------------:|:---------------------:|
| `--port="…"` (`-p`) | TCP port number with the HTTP server to check |    `8080`     |     `LISTEN_PORT`     |

<!--/GENERATED:CLI_DOCS-->

## 🦾 Contributors

I want to say a big thank you to everyone who contributed to this project:

[![contributors](https://contrib.rocks/image?repo=tarampampam/error-pages)][contributors]

[contributors]:https://github.com/tarampampam/error-pages/graphs/contributors

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

[ingress-nginx]:https://github.com/kubernetes/ingress-nginx/tree/main/charts/ingress-nginx
