# Customize error pages in your own Caddy Docker image

> [!IMPORTANT]
> Here we use an image with the `-builder` suffix. It's intended for generating error pages and already includes 
> pre-rendered templates.

To set this up, we need two things:

- A Caddyfile configuration
- A Dockerfile to build the image

Let's start with the Caddyfile:

```caddy
# File: Caddyfile

:80 {
    root * /usr/share/caddy
    file_server

    handle_errors {
        root * /usr/share/errorpages
        rewrite * /{err.status_code}.html
        file_server
    }
}
```

`handle_errors` is triggered whenever Caddy itself returns an error status (for example, a 404 from `file_server` or 
a 502 from `reverse_proxy`). The `{err.status_code}` placeholder is replaced with the actual status code, so Caddy 
will look for `/404.html`, `/502.html`, and so on in the error pages directory.

Now the Dockerfile:

```dockerfile
FROM docker.io/library/caddy:2.11-alpine

# override the default Caddy configuration
COPY --chown=root ./Caddyfile /etc/caddy/Caddyfile

# copy statically built error pages from the error-pages "builder" image
# (you can use a different template instead of `ghost` if you want)
COPY --chown=root \
     --from=ghcr.io/tarampampam/error-pages:4-builder \
     /opt/html/ghost /usr/share/errorpages
```

Next, build and run the image:

```shell
docker build --tag your-caddy:local -f ./Dockerfile .
docker run --rm -p '8081:80/tcp' your-caddy:local
```

And voilà! Let's verify that everything works as expected:

```shell
curl -s http://127.0.0.1:8081/foobar | head -n 15 # run this in another terminal

<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="robots" content="nofollow,noarchive,noindex">
  <title>404: Not Found</title>
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
