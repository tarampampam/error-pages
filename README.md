<p align="center">
  <img src="https://hsto.org/webt/rm/9y/ww/rm9ywwx3gjv9agwkcmllhsuyo7k.png" width="94" alt="" />
</p>

# HTTP's error pages

[![Release version][badge-release]][releases]
![Project language][badge-lang]
[![Build Status][badge-ci-build]][actions-page]
[![Release Status][badge-ci-release]][actions-page]
[![Coverage][badge-coverage]][coverage]
[![Image size][badge-image-size]][docker-hub]
[![License][badge-license]][license]

One day you may want to replace the standard error pages of your HTTP server with something more original and pretty. That's what this repository was created for :) It contains:

- Simple error pages generator, written on Go
- Single-page error page templates with different designs (located in the [templates](https://github.com/tarampampam/error-pages/tree/master/templates) directory)
- Fast and lightweight HTTP server
- Already generated error pages (sources can be [found here][preview-sources], the **demonstration** is always accessible [here][preview-demo])

## 🔥 Features list

- HTTP server written on Go, with the extremely fast [FastHTTP][fasthttp] under the hood
  - Respects the `Content-Type` HTTP header (and `X-Format`) value and responds with the corresponding format (supported formats is `json` and `xml`)
  - Writes logs in `json` format
  - Contains healthcheck endpoint (`/healthz`)
  - Contains metrics endpoint (`/metrics`) in Prometheus format
- Lightweight docker image _(~3.7Mb compressed size)_, distroless and uses the unleveled user by default
- [Go-template](https://pkg.go.dev/text/template) tags are allowed in the templates
- Ready for integration with [Traefik][traefik] ([error pages customization](https://doc.traefik.io/traefik/middlewares/http/errorpages/)) and [Ingress-nginx][ingress-nginx]
- Fully configurable (take a look at the [configuration file](https://github.com/tarampampam/error-pages/blob/master/error-pages.yml) and [project Wiki][wiki])
- Distributed using docker image and compiled binary files

## 🧩 Install

Download the latest binary file for your os/arch from the [releases page][releases] or use our docker image:

[![image stats](https://dockeri.co/image/tarampampam/error-pages)][docker-hub-tags]

| Registry                          | Image                             |
|-----------------------------------|-----------------------------------|
| [Docker Hub][docker-hub]          | `tarampampam/error-pages`         |
| [GitHub Container Registry][ghcr] | `ghcr.io/tarampampam/error-pages` |

> Using the `latest` tag for the docker image is highly discouraged because of possible backward-incompatible changes during **major** upgrades. Please, use tags in `X.Y.Z` format

## 🛠 Usage

Please, take a look at [our Wiki][wiki] for the common usage stories:

- [HTTP server][wiki-http-server] (routes, formats, flags and environment variables)
- [Pages generator][wiki-generator] (build your own error page set)
- [Static error pages][wiki-static-error-pages] (extract generated static error pages from the docker image)
- [Usage with nginx][wiki-usage-with-nginx] (include our error pages into an image with nginx)
- [Usage with Traekik and local Docker Compose][wiki-traefik-docker-compose] (it's a good starting point for the tests)
- [Usage with Traefik and Docker Swarm][wiki-traefik-swarm]
- [Kubernetes & ingress nginx][wiki-k8s-ingress-nginx]

[wiki]:https://github.com/tarampampam/error-pages/wiki
[wiki-http-server]:https://github.com/tarampampam/error-pages/wiki/HTTP-server
[wiki-generator]:https://github.com/tarampampam/error-pages/wiki/Generator
[wiki-static-error-pages]:https://github.com/tarampampam/error-pages/wiki/Static-error-pages
[wiki-usage-with-nginx]:https://github.com/tarampampam/error-pages/wiki/Usage-with-nginx
[wiki-traefik-swarm]:https://github.com/tarampampam/error-pages/wiki/Traefik-(docker-swarm)
[wiki-traefik-docker-compose]:https://github.com/tarampampam/error-pages/wiki/Traefik-(docker-compose)
[wiki-k8s-ingress-nginx]:https://github.com/tarampampam/error-pages/wiki/Kubernetes-&-ingress-nginx

## 🦾 Performance

Used hardware:

- Intel® Core™ i7-10510U CPU @ 1.80GHz × 8
- 16 GiB RAM

```shell
$ ulimit -aH | grep file
-f: file size (blocks)              unlimited
-c: core file size (blocks)         unlimited
-n: file descriptors                1048576
-x: file locks                      unlimited

$ wrk --version | head -n 1
wrk 4.2.0 [epoll] Copyright (C) 2012 Will Glozer

$ wrk -t12 -c400 -d30s http://127.0.0.1:8080/500.html
Running 30s test @ http://127.0.0.1:8080/500.html
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    50.06ms   61.15ms 655.79ms   85.54%
    Req/Sec     1.07k   363.14     2.40k    69.24%
  383014 requests in 30.08s, 2.14GB read
Requests/sec:  12731.07
Transfer/sec:     72.79MB
```

FS & memory usage stats during the test:

<p align="center">
  <img src="https://hsto.org/webt/dy/2e/_8/dy2e_8xkefxre7z5w7xcorjldmm.png" alt="" />
</p>

## 🪂 Templates

|       Name        |                              Preview                               |
|:-----------------:|:------------------------------------------------------------------:|
|      `ghost`      |                [![ghost][ghost-screen]][ghost-link]                |
|    `l7-light`     |           [![l7-light][l7-light-screen]][l7-light-link]            |
|     `l7-dark`     |             [![l7-dark][l7-dark-screen]][l7-dark-link]             |
|     `shuffle`     |             [![shuffle][shuffle-screen]][shuffle-link]             |
|      `noise`      |                [![noise][noise-screen]][noise-link]                |
| `hacker-terminal` | [![hacker-terminal][hacker-terminal-screen]][hacker-terminal-link] |
|      `cats`       |                 [![cats][cats-screen]][cats-link]                  |

> Note: `noise` template highly uses the CPU, be careful

[ghost-screen]:https://hsto.org/webt/oj/cl/4k/ojcl4ko_cvusy5xuki6efffzsyo.gif
[ghost-link]:https://tarampampam.github.io/error-pages/ghost/404.html
[l7-light-screen]:https://hsto.org/webt/xc/iq/vt/xciqvty-aoj-rchfarsjhutpjny.png
[l7-light-link]:https://tarampampam.github.io/error-pages/l7-light/404.html
[l7-dark-screen]:https://hsto.org/webt/s1/ih/yr/s1ihyrqs_y-sgraoimfhk6ypney.png
[l7-dark-link]:https://tarampampam.github.io/error-pages/l7-dark/404.html
[shuffle-screen]:https://hsto.org/webt/7w/rk/3m/7wrk3mrzz3y8qfqwovmuvacu-bs.gif
[shuffle-link]:https://tarampampam.github.io/error-pages/shuffle/404.html
[noise-screen]:https://hsto.org/webt/42/oq/8y/42oq8yok_i-arrafjt6hds_7ahy.gif
[noise-link]:https://tarampampam.github.io/error-pages/noise/404.html
[hacker-terminal-screen]:https://hsto.org/webt/5s/l0/p1/5sl0p1_ud_nalzjzsj5slz6dfda.gif
[hacker-terminal-link]:https://tarampampam.github.io/error-pages/hacker-terminal/404.html
[cats-screen]:https://hsto.org/webt/_g/y-/ke/_gy-keqinz-3867jbw36v37-iwe.jpeg
[cats-link]:https://tarampampam.github.io/error-pages/cats/404.html

## 🦾 Contributors

I want to say a big thank you to everyone who contributed to this project:

[![contributors](https://contrib.rocks/image?repo=tarampampam/error-pages)][contributors]

[contributors]:https://github.com/tarampampam/error-pages/graphs/contributors

## 📰 Changes log

[![Release date][badge-release-date]][releases]
[![Commits since latest release][badge-commits]][commits]

Changes log can be [found here][changelog].

## 👾 Support

[![Issues][badge-issues]][issues]
[![Issues][badge-prs]][prs]

If you find any bugs in the project, please [create an issue][new-issue] in the current repository.

## 📖 License

This is open-sourced software licensed under the [MIT License][license].

[badge-ci-build]:https://img.shields.io/github/workflow/status/tarampampam/error-pages/tests?maxAge=30&label=tests&logo=github
[badge-ci-release]:https://img.shields.io/github/workflow/status/tarampampam/error-pages/release?maxAge=30&label=release&logo=github
[badge-coverage]:https://img.shields.io/codecov/c/github/tarampampam/error-pages/master.svg?maxAge=30
[badge-release]:https://img.shields.io/github/release/tarampampam/error-pages.svg?maxAge=30
[badge-image-size]:https://img.shields.io/docker/image-size/tarampampam/error-pages/latest?maxAge=30
[badge-lang]:https://img.shields.io/github/go-mod/go-version/tarampampam/error-pages?longCache=true
[badge-license]:https://img.shields.io/github/license/tarampampam/error-pages.svg?longCache=true
[badge-release-date]:https://img.shields.io/github/release-date/tarampampam/error-pages.svg?maxAge=180
[badge-commits]:https://img.shields.io/github/commits-since/tarampampam/error-pages/latest.svg?maxAge=45
[badge-issues]:https://img.shields.io/github/issues/tarampampam/error-pages.svg?maxAge=45
[badge-prs]:https://img.shields.io/github/issues-pr/tarampampam/error-pages.svg?maxAge=45

[coverage]:https://codecov.io/gh/tarampampam/error-pages
[actions-page]:https://github.com/tarampampam/error-pages/actions
[docker-hub]:https://hub.docker.com/r/tarampampam/error-pages
[docker-hub-tags]:https://hub.docker.com/r/tarampampam/error-pages/tags
[license]:https://github.com/tarampampam/error-pages/blob/master/LICENSE
[releases]:https://github.com/tarampampam/error-pages/releases
[commits]:https://github.com/tarampampam/error-pages/commits
[changelog]:https://github.com/tarampampam/error-pages/blob/master/CHANGELOG.md
[issues]:https://github.com/tarampampam/error-pages/issues
[new-issue]:https://github.com/tarampampam/error-pages/issues/new/choose
[prs]:https://github.com/tarampampam/error-pages/pulls
[ghcr]:https://github.com/users/tarampampam/packages/container/package/error-pages

[fasthttp]:https://github.com/valyala/fasthttp
[preview-sources]:https://github.com/tarampampam/error-pages/tree/gh-pages
[preview-demo]:https://tarampampam.github.io/error-pages/
[traefik]:https://github.com/traefik/traefik
[ingress-nginx]:https://github.com/kubernetes/ingress-nginx/tree/main/charts/ingress-nginx
