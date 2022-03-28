<p align="center">
  <a href="https://github.com/tarampampam/error-pages#readme"><img src="https://socialify.git.ci/tarampampam/error-pages/image?description=1&font=Raleway&forks=1&issues=1&logo=https%3A%2F%2Fhsto.org%2Fwebt%2Frm%2F9y%2Fww%2Frm9ywwx3gjv9agwkcmllhsuyo7k.png&owner=1&pulls=1&pattern=Solid&stargazers=1&theme=Dark" alt="banner" width="100%" /></a>
</p>

<p align="center">
  <a href="#"><img src="https://img.shields.io/github/go-mod/go-version/tarampampam/error-pages?longCache=true&label=&logo=go&logoColor=white&style=flat-square" alt="" /></a>
  <a href="https://codecov.io/gh/tarampampam/error-pages"><img src="https://img.shields.io/codecov/c/github/tarampampam/error-pages/master.svg?maxAge=30&label=&logo=codecov&logoColor=white&style=flat-square" alt="" /></a>
  <a href="https://github.com/tarampampam/error-pages/actions"><img src="https://img.shields.io/github/workflow/status/tarampampam/error-pages/tests?maxAge=30&label=tests&logo=github&style=flat-square" alt="" /></a>
  <a href="https://github.com/tarampampam/error-pages/actions"><img src="https://img.shields.io/github/workflow/status/tarampampam/error-pages/release?maxAge=30&label=release&logo=github&style=flat-square" alt="" /></a>
  <a href="https://hub.docker.com/r/tarampampam/error-pages"><img src="https://img.shields.io/docker/pulls/tarampampam/error-pages.svg?maxAge=30&label=pulls&logo=docker&logoColor=white&style=flat-square" alt="" /></a>
  <a href="https://hub.docker.com/r/tarampampam/error-pages"><img src="https://img.shields.io/docker/image-size/tarampampam/error-pages/latest?maxAge=30&label=size&logo=docker&logoColor=white&style=flat-square" alt="" /></a>
  <a href="https://github.com/tarampampam/error-pages/blob/master/LICENSE"><img src="https://img.shields.io/github/license/tarampampam/error-pages.svg?maxAge=30&style=flat-square" alt="" /></a>
</p>

<p align="center"><sup>22 feb. 2022 - ⚡ Our Docker image was downloaded <strong>one MILLION times</strong> from the docker hub! ⚡</sup></p>

One day you may want to replace the standard error pages of your HTTP server with something more original and pretty. That's what this repository was created for :) It contains:

- Simple error pages generator, written in Go
- Single-page error page templates with different designs (located in the [templates](https://github.com/tarampampam/error-pages/tree/master/templates) directory)
- Fast and lightweight HTTP server
- Already generated error pages (sources can be [found here][preview-sources], the **demonstration** is always accessible [here][preview-demo])

## 🔥 Features list

- HTTP server written in Go, with the extremely fast [FastHTTP][fasthttp] under the hood
  - Respects the `Content-Type` HTTP header (and `X-Format`) value and responds with the corresponding format (supported formats are `json` and `xml`)
  - Writes logs in `json` format
  - Contains healthcheck endpoint (`/healthz`)
  - Contains metrics endpoint (`/metrics`) in Prometheus format
- Lightweight docker image _(~4.6Mb compressed size)_, distroless and uses the unleveled user by default
- [Go-template](https://pkg.go.dev/text/template) tags are allowed in the templates
- Ready for integration with [Traefik][traefik] ([error pages customization](https://doc.traefik.io/traefik/middlewares/http/errorpages/)) and [Ingress-nginx][ingress-nginx]
- Error pages can be [embedded into your own `nginx`][wiki-usage-with-nginx] docker image
- Fully configurable (take a look at the [configuration file](https://github.com/tarampampam/error-pages/blob/master/error-pages.yml) and [project Wiki][wiki])
- Distributed using docker image and compiled binary files
- Localized (🇺🇸, 🇫🇷, 🇺🇦, 🇷🇺) HTML error pages (translation process [described here](https://github.com/tarampampam/error-pages/tree/master/l10n) - other translations are welcome!)

## 🧩 Install

Download the latest binary file for your os/arch from the [releases page][releases] or use our docker image:

[![image stats](https://dockeri.co/image/tarampampam/error-pages)][docker-hub-tags]

| Registry                          | Image                             |
|-----------------------------------|-----------------------------------|
| [Docker Hub][docker-hub]          | `tarampampam/error-pages`         |
| [GitHub Container Registry][ghcr] | `ghcr.io/tarampampam/error-pages` |

> Using the `latest` tag for the docker image is highly discouraged because of possible backward-incompatible changes during **major** upgrades. Please, use tags in `X.Y.Z` format

💣 **Or** you can download **already rendered** error pages pack as a [zip][pages-pack-zip] or [tar.gz][pages-pack-tar-gz] archive.

[pages-pack-zip]:https://github.com/tarampampam/error-pages/zipball/gh-pages/
[pages-pack-tar-gz]:https://github.com/tarampampam/error-pages/tarball/gh-pages/

## 🛠 Usage

Please, take a look at [our Wiki][wiki] for the common usage stories:

- [HTTP server][wiki-http-server] (routes, formats, flags and environment variables)
- [Pages generator][wiki-generator] (build your own error page set)
- [Static error pages][wiki-static-error-pages] (extract generated static error pages from the docker image)
- [Usage with nginx][wiki-usage-with-nginx] (include our error pages into an image with nginx)
- [Usage with Traefik and local Docker Compose][wiki-traefik-docker-compose] (it's a good starting point for the tests)
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

$ docker run --rm -p "8080:8080/tcp" -e "SHOW_DETAILS=true" error-pages:local # in separate terminal

$ wrk --timeout 1s -t12 -c400 -d30s -s ./test/wrk/request.lua http://127.0.0.1:8080/
Running 30s test @ http://127.0.0.1:8080/
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    10.84ms    7.89ms 135.91ms   79.36%
    Req/Sec     3.23k   785.11     6.30k    70.04%
  1160567 requests in 30.10s, 4.12GB read
Requests/sec:  38552.04
Transfer/sec:    140.23MB
```

<details>
  <summary>FS & memory usage stats during the test</summary>

  <p align="center">
    <img src="https://hsto.org/webt/ts/w-/lz/tsw-lznvru0ngjneiimkwq7ysyc.png" alt="" />
  </p>
</details>

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
|  `lost-in-space`  |    [![lost-in-space][lost-in-space-screen]][lost-in-space-link]    |
|    `app-down`     |           [![app-down][app-down-screen]][app-down-link]            |
|   `connection`    |        [![connection][connection-screen]][connection-link]         |
|     `matrix`      |              [![matrix][matrix-screen]][matrix-link]               |

> Note: `noise` template highly uses the CPU, be careful

[ghost-screen]:https://hsto.org/webt/oj/cl/4k/ojcl4ko_cvusy5xuki6efffzsyo.gif
[ghost-link]:https://tarampampam.github.io/error-pages/ghost/404.html
[l7-light-screen]:https://hsto.org/webt/hx/ca/mm/hxcammfm7qjmogtvsjxcidgf7c8.png
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
[lost-in-space-screen]:https://hsto.org/webt/lf/ln/x8/lflnx8fuy4rofxju34ttskijdsu.gif
[lost-in-space-link]:https://tarampampam.github.io/error-pages/lost-in-space/404.html
[app-down-screen]:https://habrastorage.org/webt/j2/la/fj/j2lafjvu_xjflzrvhiixobxy_ca.png
[app-down-link]:https://tarampampam.github.io/error-pages/app-down/404.html
[connection-screen]:https://hsto.org/webt/x4/ah/jb/x4ahjboo4-arm3bxpaash_sflmw.png
[connection-link]:https://tarampampam.github.io/error-pages/connection/404.html
[matrix-screen]:https://hsto.org/webt/ng/tf/oi/ngtfoiolvmq6hf15kimcxmhprhk.gif
[matrix-link]:https://tarampampam.github.io/error-pages/matrix/404.html

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

[badge-release]:https://img.shields.io/github/release/tarampampam/error-pages.svg?maxAge=30
[badge-release-date]:https://img.shields.io/github/release-date/tarampampam/error-pages.svg?maxAge=180
[badge-commits]:https://img.shields.io/github/commits-since/tarampampam/error-pages/latest.svg?maxAge=45
[badge-issues]:https://img.shields.io/github/issues/tarampampam/error-pages.svg?maxAge=45
[badge-prs]:https://img.shields.io/github/issues-pr/tarampampam/error-pages.svg?maxAge=45

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
