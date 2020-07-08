<p align="center">
  <img src="https://hsto.org/webt/rg/ys/c3/rgysc33oc7jiufdzmwrkohpmef8.png" width="94" alt="" />
</p>

# Static error pages in a Docker container

[![Build Status][badge_build_status]][link_build_status]
[![License][badge_license]][link_license]

This repository contains a very simple generator for server error pages _(like `404: Not found`)_ and ready docker image with web server for error pages serving.

Generator ([`bin/generator.js`](./bin/generator.js)) allows you:

- Use different templates (section `templates` in configuration file)
- Generate pages with arbitrary content according to a specific template

Can be used for [Traefik error pages customization](https://docs.traefik.io/middlewares/errorpages/).

### Usage

Just execute (installed `nodejs` is required):

```bash
$ bin/generator.js -c ./configuration.json -o ./out
```

And watch into `./out` directory:

```text
./out
└── ghost
    ├── 400.html
    ├── 401.html
    ├── 403.html
    ├── 404.html
    ├── ...
    └── 505.html
```

Default configuration can be found in [`configuration.json`](./configuration.json) file.

### Docker

[![Image size][badge_size_latest]][link_docker_build]

Start image (`nginx` inside):

```bash
$ docker run --rm -p "8080:8080" tarampampam/error-pages:1.0.0
```

And open in your browser `http://127.0.0.1:8080/ghost/400.html`. Additionally, you can set "default" pages theme by passing `TEMPLATE_NAME` environment variable (eg.: `-e "TEMPLATE_NAME=ghost"`) - in this case all error pages will be accessible in root directory (eg.: `http://127.0.0.1:8080/400.html`).

Also you can use generated error pages in your own docker images:

```dockerfile
FROM nginx:1.18-alpine

COPY --from=tarampampam/error-pages:1.0.0 /opt/html/ghost /usr/share/nginx/html/error-pages
```

> [`error_page` for `nginx` configuration](http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page)

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
[link_releases]:https://github.com/tarampampam/error-pages/releases
[link_commits]:https://github.com/tarampampam/error-pages/commits
[link_changes_log]:https://github.com/tarampampam/error-pages/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/error-pages/issues
[link_pulls]:https://github.com/tarampampam/error-pages/pulls
[link_build_status]:https://travis-ci.org/tarampampam/error-pages
[link_create_issue]:https://github.com/tarampampam/error-pages/issues/new
[link_license]:https://github.com/tarampampam/error-pages/blob/master/LICENSE
[link_docker_build]:https://hub.docker.com/r/tarampampam/error-pages/
