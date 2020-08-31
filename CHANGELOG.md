# Changelog

All notable changes to this package will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog] and this project adheres to [Semantic Versioning][semver].

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
