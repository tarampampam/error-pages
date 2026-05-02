# Use Nginx as a reverse proxy with custom error pages

This guide shows how to use Nginx as a reverse proxy where any 4xx/5xx response from an upstream application is
replaced with a styled error page served by a live error-pages sidecar.

> [!NOTE]
> This approach requires no image rebuilds - the error-pages container runs alongside Nginx and serves pages on
> demand. For a static alternative that bakes pages into a custom image, see the [Nginx image guide](nginx_image.md).

We need two files: an `nginx.conf` and a `compose.yml`.

The Nginx configuration:

```nginx
# File: nginx.conf

server {
    listen      80;
    server_name test.localtest.me; # localtest.me resolves to 127.0.0.1 via public DNS

    # without this, Nginx forwards upstream 4xx/5xx responses to the client as-is;
    # this directive makes Nginx intercept them and apply the error_page rules below
    proxy_intercept_errors on;

    # nginx has no range syntax for error_page - codes must be listed individually
    error_page 400 401 403 404 405 408 409 410 429 500 502 503 504 /_error-proxy;

    location = /_error-proxy {
        internal; # not reachable from outside; only triggered by error_page above
        proxy_set_header X-Code $status; # $status holds the intercepted upstream code
        proxy_pass      http://error-pages:8080;
    }

    location / {
        proxy_pass http://httpbin:8080;
    }
}
```

The Compose file:

```yaml
# yaml-language-server: $schema=https://cdn.jsdelivr.net/gh/compose-spec/compose-spec@master/schema/compose-spec.json
# file: compose.yml

services:
  nginx:
    image: docker.io/library/nginx:1.29-alpine
    ports:
      - "80:80/tcp"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      # do not start Nginx until error-pages is up and passing its health check,
      # ensuring the error handler backend is ready before any traffic arrives
      error-pages: {condition: service_healthy}

  error-pages:
    image: ghcr.io/tarampampam/error-pages:4
    environment:
      TEMPLATE_NAME: l7

  httpbin:
    # go-httpbin: /status/{code} returns the requested HTTP status code - useful for testing
    image: ghcr.io/mccutchen/go-httpbin:2.22
```

Place both files in the same directory, then run:

```shell
docker compose up
```

Now you can verify that everything works as expected:

```shell
# httpbin responds normally - error-pages is not involved
curl -s http://test.localtest.me/get

{
  "args": {},
  "headers": {
    "Accept": [
      "*/*"
    ],
    "Host": [
      "httpbin:8080"
    ],
    "User-Agent": [
      "curl/8.11.1"
    ]
  },
  "method": "GET",
  "origin": "172.20.0.4",
  "url": "http://httpbin:8080/get"
}
```

```shell
# httpbin returns 404; Nginx intercepts it and serves the styled error page instead
curl -s -H "Accept: text/html" http://test.localtest.me/status/404 | head -n 15

<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="robots" content="nofollow,noarchive,noindex">
  <title>Not Found</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <!--  -->
  <meta name="title" content="404: Not Found">
  <meta name="description" content="The server can not find the requested page">
  <meta property="og:title" content="404: Not Found">
  <meta property="og:description" content="The server can not find the requested page">
  <meta property="twitter:title" content="404: Not Found">
  <meta property="twitter:description" content="The server can not find the requested page">
  <style>
```

```shell
# the same response, but in JSON format - the error page template can also generate JSON
curl -s -H "Accept: application/json" http://test.localtest.me/status/404

{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}
```

The same applies to any other code listed in `error_page` - try `/status/503`, `/status/429`, etc. For the
upstream-unreachable case, Nginx will generate a `502 Bad Gateway`, which is also intercepted and styled:

```shell
# stop httpbin to trigger a 502 Bad Gateway error in Nginx
docker stop $(docker ps -aq --filter "name=httpbin")

curl -s -H "Accept: application/json" http://test.localtest.me/status/404 | head -n 15

{
  "error": true,
  "code": 502,
  "message": "Bad Gateway",
  "description": "The server received an invalid response from the upstream server"
}
```
