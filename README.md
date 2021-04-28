<p align="center">
  <img src="https://hsto.org/webt/rm/9y/ww/rm9ywwx3gjv9agwkcmllhsuyo7k.png" width="94" alt="" />
</p>

# HTTP's error pages in Docker image

[![Build Status][badge_build_status]][link_build_status]
[![Image size][badge_size_latest]][link_docker_hub]
[![Docker Pulls][badge_docker_pulls]][link_docker_hub]
[![License][badge_license]][link_license]

This repository contains:

- A very simple [generator](generator/generator.js) _(`nodejs`)_ for HTTP error pages _(like `404: Not found`)_ with different templates supports
- Dockerfile for docker image ([docker hub][link_docker_hub], [ghcr.io][link_ghcr]) with generated pages and `nginx` as a web server

**Can be used for [Traefik error pages customization](https://docs.traefik.io/middlewares/errorpages/)**.

## Demo

Generated pages (from the latest release) always **[accessible here][link_gh_pages]** _(live preview)_.

## Templates

       Name       | Preview
:---------------: | :-----:
`ghost`           | [![ghost](https://hsto.org/webt/oj/cl/4k/ojcl4ko_cvusy5xuki6efffzsyo.gif)](https://tarampampam.github.io/error-pages/ghost/404.html)
`l7-light`        | [![l7-light](https://hsto.org/webt/xc/iq/vt/xciqvty-aoj-rchfarsjhutpjny.png)](https://tarampampam.github.io/error-pages/l7-light/404.html)
`l7-dark`         | [![l7-dark](https://hsto.org/webt/s1/ih/yr/s1ihyrqs_y-sgraoimfhk6ypney.png)](https://tarampampam.github.io/error-pages/l7-dark/404.html)
`shuffle`         | [![shuffle](https://hsto.org/webt/7w/rk/3m/7wrk3mrzz3y8qfqwovmuvacu-bs.gif)](https://tarampampam.github.io/error-pages/shuffle/404.html)
`noise`           | [![noise](https://hsto.org/webt/42/oq/8y/42oq8yok_i-arrafjt6hds_7ahy.gif)](https://tarampampam.github.io/error-pages/noise/404.html)
`hacker-terminal` | [![hacker-terminal](https://hsto.org/webt/5s/l0/p1/5sl0p1_ud_nalzjzsj5slz6dfda.gif)](https://tarampampam.github.io/error-pages/hacker-terminal/404.html)

> Note: `noise` template highly uses the CPU, be careful

## Usage

Generated error pages in our [docker image][link_docker_hub] permanently located in directory `/opt/html/%TEMPLATE_NAME%`. `nginx` in a container listen for `8080` (`http`) port.

#### Supported environment variables

Name            | Description
--------------- | -----------
`TEMPLATE_NAME` | (`ghost` by default) "default" pages template _(allows to use error pages without passing theme name in URL - `http://127.0.0.1/500.html` instead `http://127.0.0.1/ghost/500.html`)_

Also, you can use a special template name `random` - in this case template will be selected randomly.

### Ready docker image

Execute in your shell:

```bash
$ docker run --rm -p "8082:8080" tarampampam/error-pages:X.X.X
```

> Important notice: do **not** use the `latest` image tag _(this is bad practice)_. Use versioned tag (like `1.2.3`) instead.

And open in your browser `http://127.0.0.1:8082/ghost/400.html`.

### Custom error pages for your image with [nginx][link_nginx]

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

FROM nginx:1.18-alpine

COPY --chown=nginx \
     ./nginx.conf /etc/nginx/conf.d/default.conf
COPY --chown=nginx \
     --from=tarampampam/error-pages:1.7.0 \
     /opt/html/ghost /usr/share/nginx/errorpages/_error-pages
```

```shell
$ docker build --tag your-nginx:local -f ./Dockerfile .
```

> More info about `error_page` directive can be [found here](http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page).

### Custom error pages for [Traefik][link_traefik]

Simple traefik (tested on `v2.4.8`) service configuration for usage in [docker swarm][link_swarm] (**change with your needs**):

```yaml
version: '3.4'

services:
  error-pages:
    image: tarampampam/error-pages:1.7.0
    environment:
      TEMPLATE_NAME: l7-dark
    networks:
      - traefik-public
    deploy:
      placement:
        constraints:
          - node.role == worker
      labels:
        traefik.enable: 'true'
        traefik.docker.network: traefik-public
        # use as "fallback" for any non-registered services (with priority below normal)
        traefik.http.routers.error-pages-router.rule: HostRegexp(`{host:.+}`)
        traefik.http.routers.error-pages-router.priority: 10
        # should say that all of your services work on https
        traefik.http.routers.error-pages-router.tls: 'true'
        traefik.http.routers.error-pages-router.entrypoints: https
        traefik.http.routers.error-pages-router.middlewares: error-pages-middleware@docker
        traefik.http.services.error-pages-service.loadbalancer.server.port: 8080
        # "errors" middleware settings
        traefik.http.middlewares.error-pages-middleware.errors.status: 400-599
        traefik.http.middlewares.error-pages-middleware.errors.service: error-pages-service@docker
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
        traefik.enable: 'true'
        traefik.docker.network: traefik-public
        traefik.http.routers.another-service.rule: Host(`subdomain.example.com`)
        traefik.http.routers.another-service.tls: 'true'
        traefik.http.routers.another-service.entrypoints: https
        # next line is important
        traefik.http.routers.another-service.middlewares: error-pages-middleware@docker
        traefik.http.services.another-service.loadbalancer.server.port: 80

networks:
  traefik-public:
    external: true
```

## Development

> For project development we use `docker-ce` + `docker-compose`. Make sure you have them installed.

Install "generator" dependencies:

```bash
$ make install
```

If you want to generate error pages on your machine _(after that look into the output directory)_:

```bash
$ make gen
```

If you want to preview the pages using the Docker image:

```bash
$ make preview
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

[badge_build_status]:https://img.shields.io/github/workflow/status/tarampampam/error-pages/tests/master
[badge_release_date]:https://img.shields.io/github/release-date/tarampampam/error-pages.svg?style=flat-square&maxAge=180
[badge_commits_since_release]:https://img.shields.io/github/commits-since/tarampampam/error-pages/latest.svg?style=flat-square&maxAge=180
[badge_issues]:https://img.shields.io/github/issues/tarampampam/error-pages.svg?style=flat-square&maxAge=180
[badge_pulls]:https://img.shields.io/github/issues-pr/tarampampam/error-pages.svg?style=flat-square&maxAge=180
[badge_license]:https://img.shields.io/github/license/tarampampam/error-pages.svg?longCache=true
[badge_size_latest]:https://img.shields.io/docker/image-size/tarampampam/error-pages/latest?maxAge=30
[badge_docker_pulls]:https://img.shields.io/docker/pulls/tarampampam/error-pages.svg
[link_releases]:https://github.com/tarampampam/error-pages/releases
[link_commits]:https://github.com/tarampampam/error-pages/commits
[link_changes_log]:https://github.com/tarampampam/error-pages/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/error-pages/issues
[link_pulls]:https://github.com/tarampampam/error-pages/pulls
[link_build_status]:https://travis-ci.org/tarampampam/error-pages
[link_create_issue]:https://github.com/tarampampam/error-pages/issues/new
[link_license]:https://github.com/tarampampam/error-pages/blob/master/LICENSE
[link_docker_hub]:https://hub.docker.com/r/tarampampam/error-pages/
[link_ghcr]:https://github.com/users/tarampampam/packages/container/package/error-pages
[link_nginx]:http://nginx.org/
[link_traefik]:https://docs.traefik.io/
[link_swarm]:https://docs.docker.com/engine/swarm/
[link_gh_pages]:https://tarampampam.github.io/error-pages/
