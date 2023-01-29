# Changelog

All notable changes to this package will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog] and this project adheres to [Semantic Versioning][semver].

## v2.20.0

### Changed

- `version` subcommand replaced by `--version` flag [#163]
- `--config-file` flag is not global anymore (use `error-pages (serve|build) --config-file ...` instead of `error-pages --config-file ... (serve|build) ...`) [#163]
- Flags `--verbose`, `--debug` and `--log-json` are deprecated, use `--log-level` and `--log-format` instead [#163]

### Added

- Possibility to use custom env variables in templates [#164], [#165]

[#164]:https://github.com/tarampampam/error-pages/issues/164
[#165]:https://github.com/tarampampam/error-pages/pull/165
[#163]:https://github.com/tarampampam/error-pages/pull/163

## v2.19.0

### Changed

- Go updated from `1.18` up to `1.19`

### Added

- Error pages now translated into Chinese 🇨🇳 [#147]

[#147]:https://github.com/tarampampam/error-pages/pull/147

## v2.18.0

### Changed

- Replaced `fonts.googleapis.com` by `fonts.bunny.net` regarding GDPR compliance [#131]

[#131]:https://github.com/tarampampam/error-pages/pull/131

## v2.17.0

### Added

- Error pages now translated into Spanish 🇪🇸 [#124]

[#124]:https://github.com/tarampampam/error-pages/pull/124

## v2.16.0

### Added

- Error pages are now translated into German 🇩🇪 [#115]

[#115]:https://github.com/tarampampam/error-pages/pull/115

## v2.15.0

### Added

- Error pages now translated into Dutch 🇳🇱 [#104]

[#104]:https://github.com/tarampampam/error-pages/pull/104

## v2.14.0

### Added

- Error pages now translated into Portuguese 🇵🇹 [#103]

### Changed

- Go updated from `1.18.0` up to `1.18.1`

[#103]:https://github.com/tarampampam/error-pages/pull/103

## v2.13.0

### Added

- Possibility to disable error pages auto-localization (using `--disable-l10n` flag for the `serve` & `build` commands or environment variable `DISABLE_L10N`) [#91]

### Fixed

- User UID/GID changed to the numeric values in the dockerfile [#92]

[#92]:https://github.com/tarampampam/error-pages/issues/92
[#91]:https://github.com/tarampampam/error-pages/issues/91

## v2.12.1

### Fixed

- Fix translation 🇫🇷 [#86]

[#85]:https://github.com/tarampampam/error-pages/pull/86

## v2.12.0

### Changed

- Error pages now translated into 🇫🇷 [#82]

[#82]:https://github.com/tarampampam/error-pages/pull/82

## v2.11.0

### Added

- Template `matrix` [#81]

### Fixed

- Localization mistakes [#81]

[#81]:https://github.com/tarampampam/error-pages/pull/81

## v2.10.1

### Fixed

- Template `shuffle`
- Localization mistakes

## v2.10.0

### Changed

- Error pages now translated into 🇺🇦 and 🇷🇺 languages [#80]

[#80]:https://github.com/tarampampam/error-pages/pull/80

## v2.9.0

### Added

- Template `connection` [#79]

[#79]:https://github.com/tarampampam/error-pages/pull/79

## v2.8.1

### Fixed

- Dark mode for `app-down` template

### Changed

- The index page for built error pages now supports a dark theme

## v2.8.0

### Added

- Template `app-down` [#74]

### Changed

- Go updated from `1.17.6` up to `1.18.0`

[#74]:https://github.com/tarampampam/error-pages/pull/74

## v2.7.0

### Changed

- Logs includes request/response headers now [#67]

### Added

- Possibility to proxy HTTP headers from the requests to the responses (can be enabled using `--proxy-headers` flag for the `serve` command or environment variable `PROXY_HTTP_HEADERS`, headers list should be comma-separated) [#67]
- Template `lost-in-space` [#68]

### Fixed

- Template `l7-light` uses the dark colors in the browsers with the preferred dark theme

[#67]:https://github.com/tarampampam/error-pages/pull/67
[#68]:https://github.com/tarampampam/error-pages/pull/68

## v2.6.0

### Added

- Possibility to change the template to the random once a day using "special" template name `random-daily` (or hourly, using `random-hourly`) [#48]

[#48]:https://github.com/tarampampam/error-pages/issues/48

## v2.5.0

### Changed

- Go updated from `1.17.5` up to `1.17.6`

### Added

- `Host` and `X-Forwarded-For` Header to error pages [#61]

### Fixed

- Performance issue, that affects template rendering. Now templates are cached in memory (for 2 seconds), and it has improved performance by more than 200% [#60]

[#60]:https://github.com/tarampampam/error-pages/pull/60
[#61]:https://github.com/tarampampam/error-pages/pull/61

## v2.4.0

### Changed

- It is now possible to use [golang-tags of templates](https://pkg.go.dev/text/template) in error page templates and formatted (`json`, `xml`) responses [#49]
- Health-check route become `/healthz` (instead `/health/live`, previous route marked as deprecated) [#49]

### Added

- The templates contain details block now (can be enabled using `--show-details` flag for the `serve` command or environment variable `SHOW_DETAILS=true`) [#49]
- Formatted response templates (`json`, `xml`) - the server responds with a formatted response depending on the `Content-Type` (and `X-Format`) request header value [#49]
- HTTP header `X-Robots-Tag: noindex` for the error pages [#49]
- Possibility to pass the needed error page code using `X-Code` HTTP header [#49]
- Possibility to integrate with [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) [#49]
- Metrics HTTP endpoint `/metrics` in prometheus format [#54]

### Fixed

- Potential race condition (in the `pick.StringsSlice` struct) [#49]

[#54]:https://github.com/tarampampam/error-pages/pull/54
[#49]:https://github.com/tarampampam/error-pages/pull/49

## v2.3.0

### Added

- Flag `--default-http-code` for the `serve` subcommand (`404` is used by default instead of `200`, environment name `DEFAULT_HTTP_CODE`) [#41]

### Changed

- Go updated from `1.17.1` up to `1.17.5`

[#41]:https://github.com/tarampampam/error-pages/issues/41

## v2.2.0

### Added

- Template `cats` [#31]

[#31]:https://github.com/tarampampam/error-pages/pull/31

## v2.1.0

### Added

- `referer` field in access log records
- Flag `--default-error-page` for the `serve` subcommand (`404` is used by default, environment name `DEFAULT_ERROR_PAGE`)

### Changed

- The source code has been refactored
- The index page (`/`) now returns the error page with a code, declared using `--default-error-page` flag (HTTP code 200, when a page code exists)

## v2.0.0

### Changed

- Application rewritten in Go

## v1.8.0

### Added

- Nginx health-check endpoint (`/health/live`) and dockerfile `HEALTHCHECK` to utilise (thx [@modem7](https://github.com/modem7)) [#22], [#23]

[#22]:https://github.com/tarampampam/error-pages/pull/22
[#23]:https://github.com/tarampampam/error-pages/pull/23

## v1.7.2

### Changed

- Nginx updated up to `1.21` (from `1.19`)

## v1.7.1

### Fixed

- Random template selecting (thx [@xpliz](https://github.com/xpliz)) [#12]

[#12]:https://github.com/tarampampam/error-pages/pull/12

## v1.7.0

### Added

- Template `hacker-terminal` [#13]
- HTML comments with error code and description into each template (header and footer, it seems more readable for curl usage)

[#10]:https://github.com/tarampampam/error-pages/pull/13

## v1.6.0

### Added

- Template `noise` [#10]

### Fixed

- File permissions in docker image

[#10]:https://github.com/tarampampam/error-pages/issues/10

## v1.5.0

### Changed

- Repository files structure
- Nginx updated from `1.18` up to `1.19` in docker image
- Docker image now uses default `nginx` entrypoint scripts and command

### Added

- Support for `linux/arm64/v8`, `linux/arm/v6` and `linux/arm/v7` platforms for docker image
- Random template selecting (use `random` as a template name) for docker image

## v1.4.0

### Added

- Template `shuffle` [#4]

[#4]:https://github.com/tarampampam/error-pages/issues/4

## v1.3.1

### Fixed

- `can't create directory '/opt/html/nginx-error-pages'` error [#3]

[#3]:https://github.com/tarampampam/error-pages/issues/3

## v1.3.0

### Added

- `418` status code error page
- Set `server_tokens off;` in `nginx` server configuration

## v1.2.0

### Fixed

- By default `nginx` in docker container returns 404 http code instead 200 when `/` requested

### Changed

- Default value for `TEMPLATE_NAME` is `ghost` now

### Removed

- Environment variable `DEFAULT_ERROR_CODE` support in docker image

### Added

- Templates `l7-light` and `l7-dark`

## v1.1.0

### Added

- Environment variable `DEFAULT_ERROR_CODE` support in docker image

## v1.0.1

### Changed

- Repository (not docker image) renamed from `error-pages-docker` to `error-pages`
- `configuration.json` renamed to `config.json`
- Makefile contains new targets (`install`, `gen`, `preview`)
- Generator logging messages

### Added

- `docker-compose` for development

### Fixed

- Readme file content [#1]

[#1]:https://github.com/tarampampam/error-pages/issues/1

## v1.0.0

### Changed

- First project release

[keepachangelog]:https://keepachangelog.com/en/1.0.0/
[semver]:https://semver.org/spec/v2.0.0.html
