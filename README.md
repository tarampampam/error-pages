<p align="center">
  <img src="https://hsto.org/webt/rm/9y/ww/rm9ywwx3gjv9agwkcmllhsuyo7k.png" width="94" alt="" />
</p>

# HTTP's error pages

[![Release version][badge_release_version]][link_releases]
![Project language][badge_language]
[![Build Status][badge_build]][link_build]
[![Release Status][badge_release]][link_build]
[![Coverage][badge_coverage]][link_coverage]
[![Image size][badge_size_latest]][link_docker_hub]
[![License][badge_license]][link_license]

One day you may want to replace the standard error pages of your HTTP server with something more original and pretty. That's what this repository was created for :) It contains:

- Simple error pages generator, written on Go
- Single-page error page templates with different designs (located in the [templates](templates) directory)
- Fast and lightweight HTTP server (written on Go also, with the [FastHTTP][fasthttp] under the hood)
- HTTP server respects the `Content-Type` HTTP header (and `X-Format`) value and responds with the corresponding format
- Already generated error pages (sources can be [found here][preview-sources], the **demonstration** is always accessible [here][preview-demo])
- Lightweight docker image _(~3.7Mb compressed size)_ with all the things described above
- Ready to integrate with the Traefik and Ingress-nginx

Also, this project can be used for the [**Traefik** error pages customization](https://doc.traefik.io/traefik/middlewares/http/errorpages/).

<p align="center">
  <img src="https://hsto.org/webt/bc/bt/9i/bcbt9i3jyvozequr1e4maz7i2q8.png" alt="" />
</p>

## Installing

Download the latest binary file for your os/arch from the [releases page][link_releases] or use our docker image:

| Registry                                   | Image                             |
|--------------------------------------------|-----------------------------------|
| [Docker Hub][link_docker_hub]              | `tarampampam/error-pages`         |
| [GitHub Container Registry][link_ghcr]     | `ghcr.io/tarampampam/error-pages` |

> Using the `latest` tag for the docker image is highly discouraged because of possible backward-incompatible changes during **major** upgrades. Please, use tags in `X.Y.Z` format

To watch the docker image content you can use the [dive](https://github.com/wagoodman/dive):

```bash
$ docker run --rm -it \
    -v "/var/run/docker.sock:/var/run/docker.sock:ro" \
    wagoodman/dive:latest \
      tarampampam/error-pages:latest
```

<details>
  <summary>Dive screenshot</summary>

<p align="center">
  <img src="https://hsto.org/webt/mi/ak/uf/miakufsh2ibxtsa1nomfhyqombi.png" alt="" />
</p>
</details>

## Templates

|       Name        |                                                                         Preview                                                                          |
|:-----------------:|:--------------------------------------------------------------------------------------------------------------------------------------------------------:|
|      `ghost`      |           [![ghost](https://hsto.org/webt/oj/cl/4k/ojcl4ko_cvusy5xuki6efffzsyo.gif)](https://tarampampam.github.io/error-pages/ghost/404.html)           |
|    `l7-light`     |        [![l7-light](https://hsto.org/webt/xc/iq/vt/xciqvty-aoj-rchfarsjhutpjny.png)](https://tarampampam.github.io/error-pages/l7-light/404.html)        |
|     `l7-dark`     |         [![l7-dark](https://hsto.org/webt/s1/ih/yr/s1ihyrqs_y-sgraoimfhk6ypney.png)](https://tarampampam.github.io/error-pages/l7-dark/404.html)         |
|     `shuffle`     |         [![shuffle](https://hsto.org/webt/7w/rk/3m/7wrk3mrzz3y8qfqwovmuvacu-bs.gif)](https://tarampampam.github.io/error-pages/shuffle/404.html)         |
|      `noise`      |           [![noise](https://hsto.org/webt/42/oq/8y/42oq8yok_i-arrafjt6hds_7ahy.gif)](https://tarampampam.github.io/error-pages/noise/404.html)           |
| `hacker-terminal` | [![hacker-terminal](https://hsto.org/webt/5s/l0/p1/5sl0p1_ud_nalzjzsj5slz6dfda.gif)](https://tarampampam.github.io/error-pages/hacker-terminal/404.html) |
|      `cats`       |           [![cats](https://hsto.org/webt/_g/y-/ke/_gy-keqinz-3867jbw36v37-iwe.jpeg)](https://tarampampam.github.io/error-pages/cats/404.html)            |

> Note: `noise` template highly uses the CPU, be careful

## Usage

All the examples below will use a docker image with the application, but you can also use a binary. By the way, our docker image uses the **unleveled user** by default and **distroless**.

<details>
  <summary><strong>HTTP server</strong></summary>

As mentioned above - our application can be run as an HTTP server. It only needs to specify the path to the configuration file (it does not need statically generated error pages). The server uses [FastHTTP][fasthttp] and stores all necessary data in memory - so it does not use the file system and very fast. Oh yes, the image with the app also contains a configured **healthcheck** and **logs in JSON** format :)

For the HTTP server running execute in your terminal:

```bash
$ docker run --rm \
    -p "8080:8080/tcp" \
    -e "TEMPLATE_NAME=random" \
    tarampampam/error-pages
```

And open [`http://127.0.0.1:8080/404.html`](http://127.0.0.1:8080/404.html) in your favorite browser. Error pages are accessible by the following URLs: `http://127.0.0.1:8080/{page_code}.html`. Another way is to pass the needed error page code (and HTTP response code) using the HTTP header `X-Code` when requesting an index page.

Additionally, the server respects the `Content-Type` HTTP header (and `X-Format`) value and responds with the corresponding format instead of the HTML-formatted response only. The `xml` and `json` content types (formats) are allowed at this moment, and its format can be fully customized using a configuration file!

For the integration with [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) you can start the server with the flag `--show-details` (environment variable `SHOW_DETAILS=true`) - in this case, the error pages (`json`, `xml` responses too) will contain additional information that passed from the upstream reverse proxy!

Environment variable `TEMPLATE_NAME` should be used for the theme switching (supported templates are described below).

> **Cheat**: you can use `random` (to use the randomized theme on server start) or `i-said-random` (to use the randomized template on **each request**)

To see the help run the following command:

```bash
$ docker run --rm tarampampam/error-pages serve --help
```
</details>

<details>
  <summary><strong>Generator</strong></summary>

Create a config file (`error-pages.yml`) with the following content:

```yaml
templates:
  - path: ./foo.html # Template name "foo" (same as file name),
                     # content located in the file "foo.html"
  - name: bar # Template name "bar", its content is described below:
    content: "Error {{ code }}: {{ message }} ({{ description }})"

pages:
  400:
    message: Bad Request
    description: The server did not understand the request

  401:
    message: Unauthorized
    description: The requested page needs a username and a password
```

Template file `foo.html`:

```html
<html>
<title>{{ code }}</title>
<body>
    <h1>{{ message }}: {{ description }}</h1>
</body>
</html>
```

And run the generator:

```bash
$ docker run --rm \
    -v "$(pwd):/opt:rw" \
    -u "$(id -u):$(id -g)" \
    tarampampam/error-pages build --config-file ./error-pages.yml ./out

$ tree
.
├── error-pages.yml
├── foo.html
└── out
    ├── bar
    │   ├── 400.html
    │   └── 401.html
    └── foo
        ├── 400.html
        └── 401.html

3 directories, 6 files

$ cat ./out/foo/400.html
<html>
<title>400</title>
<body>
    <h1>Bad Request: The server did not understand the request</h1>
</body>
</html>

$ cat ./out/bar/400.html
Error 400: Bad Request (The server did not understand the request)
```

To see the usage help run the following command:

```bash
$ docker run --rm tarampampam/error-pages build --help
```
</details>

<details>
  <summary><strong>Static error pages</strong></summary>

You may want to use the generated error pages somewhere else, and you can simply extract them from the docker image to your local directory for this purpose:

```bash
$ docker create --name error-pages tarampampam/error-pages
$ docker cp error-pages:/opt/html ./out
$ docker rm -f error-pages
$ ls ./out
ghost  hacker-terminal  index.html  l7-dark  l7-light  noise  shuffle
$ tree
.
└── out
    ├── ghost
    │   ├── 400.html
    │   ├── ...
    │   └── 505.html
    ├── hacker-terminal
    │   ├── 400.html
    │   ├── ...
    │   └── 505.html
    ├── index.html
    ├── l7-dark
    │   ├── 400.html
    │   ├── ...
    ...
```

Or inside another docker image:

```dockerfile
FROM alpine:latest

COPY --from=tarampampam/error-pages /opt/html /error-pages

RUN ls -l /error-pages
```

```bash
$ docker build --rm .

...
Step 3/3 : RUN ls -l /error-pages
 ---> Running in 30095dc344a9
total 12
drwxr-xr-x    2 root     root           326 Sep 29 15:44 ghost
drwxr-xr-x    2 root     root           326 Sep 29 15:44 hacker-terminal
-rw-r--r--    1 root     root         11241 Sep 29 15:44 index.html
drwxr-xr-x    2 root     root           326 Sep 29 15:44 l7-dark
drwxr-xr-x    2 root     root           326 Sep 29 15:44 l7-light
drwxr-xr-x    2 root     root           326 Sep 29 15:44 noise
drwxr-xr-x    2 root     root           326 Sep 29 15:44 shuffle
```
</details>

<details>
  <summary><strong>Custom error pages for your image with nginx</strong></summary>

You can build your own docker image with `nginx` and our error pages:

```nginx
# File `nginx.conf`

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

```dockerfile
# File `Dockerfile`

FROM nginx:1.23-alpine

COPY --chown=nginx \
     ./nginx.conf /etc/nginx/conf.d/default.conf
COPY --chown=nginx \
     --from=tarampampam/error-pages:2.3.0 \
     /opt/html/ghost /usr/share/nginx/errorpages/_error-pages
```

```shell
$ docker build --tag your-nginx:local -f ./Dockerfile .
```

> More info about `error_page` directive can be [found here](http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page).
</details>

## Custom error pages for [Traefik][link_traefik]

Simple traefik (tested on `v2.5.3`) service configuration for usage in [docker swarm][link_swarm] (**change with your needs**):

```yaml
version: '3.8'

services:
  error-pages:
    image: tarampampam/error-pages:2.3.0
    environment:
      TEMPLATE_NAME: l7-dark
    networks:
      - traefik-public
    deploy:
      placement:
        constraints:
          - node.role == worker
      labels:
        traefik.enable: true
        traefik.docker.network: traefik-public
        # use as "fallback" for any non-registered services (with priority below normal)
        traefik.http.routers.error-pages-router.rule: HostRegexp(`{host:.+}`)
        traefik.http.routers.error-pages-router.priority: 10
        # should say that all of your services work on https
        traefik.http.routers.error-pages-router.tls: 'true'
        traefik.http.routers.error-pages-router.entrypoints: https
        traefik.http.routers.error-pages-router.middlewares: error-pages-middleware
        traefik.http.services.error-pages-service.loadbalancer.server.port: 8080
        # "errors" middleware settings
        traefik.http.middlewares.error-pages-middleware.errors.status: 400-599
        traefik.http.middlewares.error-pages-middleware.errors.service: error-pages-service
        traefik.http.middlewares.error-pages-middleware.errors.query: /{status}.html

  any-another-http-service:
    image: nginx:alpine
    networks:
      - traefik-public
    deploy:
      placement:
        constraints:
          - node.role == worker
      labels:
        traefik.enable: true
        traefik.docker.network: traefik-public
        traefik.http.routers.another-service.rule: Host(`subdomain.example.com`)
        traefik.http.routers.another-service.tls: 'true'
        traefik.http.routers.another-service.entrypoints: https
        # next line is important
        traefik.http.routers.another-service.middlewares: error-pages-middleware
        traefik.http.services.another-service.loadbalancer.server.port: 80

networks:
  traefik-public:
    external: true
```

### Docker-compose + Traefik

```yaml
# docker-compose.yml
version: '3.8'

services:
  traefik:
    image: traefik:v2.5
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

  error-pages:
    image: tarampampam/error-pages:2.3.0
    environment:
      TEMPLATE_NAME: l7-dark # set the error pages template
    labels:
      traefik.enable: true
      # use as "fallback" for any NON-registered services (with priority below normal)
      traefik.http.routers.error-pages-router.rule: HostRegexp(`{host:.+}`)
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
    depends_on:
      - traefik

  nginx-or-any-another-service:
    image: nginx:1.21-alpine
    labels:
      traefik.enable: true
      traefik.http.routers.test-service.rule: Host(`test.localtest.me`)
      traefik.http.routers.test-service.entrypoints: web
      traefik.http.routers.test-service.middlewares: error-pages-middleware
```

After executing `docker-compose up` you can:

- Open Traefik dashboard [on a domain `traefik.localtest.me`](http://traefik.localtest.me/dashboard/#/)
- [View customized error pages on the Traefik dashboard](http://traefik.localtest.me/foobar404)
- Open nginx index page [on a domain `test.localtest.me`](http://test.localtest.me/)
- View customized error pages for non-existing [pages](http://test.localtest.me/404) and [domains](http://404.localtest.me/)

### Kubernetes + ingress-nginx

This can be configured to work with the [ingress-nginx](https://github.com/kubernetes/ingress-nginx/tree/main/charts/ingress-nginx) helm chart in Kubernetes.

1. Set the `custom-http-errors` config value
2. Enable default backend
3. Set the default backend image

```yaml
controller:
  config:
    custom-http-errors: >-
      401,403,404,500,501,502,503
defaultBackend:
  enabled: true
  image:
    repository: ghcr.io/tarampampam/error-pages
    tag: latest
  # optional: change the default theme
  extraEnvs:
  - name: TEMPLATE_NAME
    value: l7-dark
```

## Changes log

[![Release date][badge_release_date]][link_releases]
[![Commits since latest release][badge_commits_since_release]][link_commits]

Changes log can be [found here][link_changes_log].

## Support

[![Issues][badge_issues]][link_issues]
[![Issues][badge_pulls]][link_pulls]

If you will find any package errors, please, [make an issue][link_create_issue] in current repository.

## License

This is open-sourced software licensed under the [MIT License][link_license].

[badge_build]:https://img.shields.io/github/workflow/status/tarampampam/error-pages/tests?maxAge=30&label=tests&logo=github
[badge_release]:https://img.shields.io/github/workflow/status/tarampampam/error-pages/release?maxAge=30&label=release&logo=github
[badge_coverage]:https://img.shields.io/codecov/c/github/tarampampam/error-pages/master.svg?maxAge=30
[badge_release_version]:https://img.shields.io/github/release/tarampampam/error-pages.svg?maxAge=30
[badge_size_latest]:https://img.shields.io/docker/image-size/tarampampam/error-pages/latest?maxAge=30
[badge_language]:https://img.shields.io/github/go-mod/go-version/tarampampam/error-pages?longCache=true
[badge_license]:https://img.shields.io/github/license/tarampampam/error-pages.svg?longCache=true
[badge_release_date]:https://img.shields.io/github/release-date/tarampampam/error-pages.svg?maxAge=180
[badge_commits_since_release]:https://img.shields.io/github/commits-since/tarampampam/error-pages/latest.svg?maxAge=45
[badge_issues]:https://img.shields.io/github/issues/tarampampam/error-pages.svg?maxAge=45
[badge_pulls]:https://img.shields.io/github/issues-pr/tarampampam/error-pages.svg?maxAge=45

[link_coverage]:https://codecov.io/gh/tarampampam/error-pages
[link_build]:https://github.com/tarampampam/error-pages/actions
[link_docker_hub]:https://hub.docker.com/r/tarampampam/error-pages
[link_docker_tags]:https://hub.docker.com/r/tarampampam/error-pages/tags
[link_license]:https://github.com/tarampampam/error-pages/blob/master/LICENSE
[link_releases]:https://github.com/tarampampam/error-pages/releases
[link_commits]:https://github.com/tarampampam/error-pages/commits
[link_changes_log]:https://github.com/tarampampam/error-pages/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/error-pages/issues
[link_create_issue]:https://github.com/tarampampam/error-pages/issues/new/choose
[link_pulls]:https://github.com/tarampampam/error-pages/pulls
[link_ghcr]:https://github.com/users/tarampampam/packages/container/package/error-pages

[fasthttp]:https://github.com/valyala/fasthttp
[preview-sources]:https://github.com/tarampampam/error-pages/tree/gh-pages
[preview-demo]:https://tarampampam.github.io/error-pages/

[link_nginx]:http://nginx.org/
[link_traefik]:https://docs.traefik.io/
[link_swarm]:https://docs.docker.com/engine/swarm/
[link_gh_pages]:https://tarampampam.github.io/error-pages/
