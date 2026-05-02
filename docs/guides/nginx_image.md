# Customize error pages in your own Nginx Docker image

> [!IMPORTANT]
> This example uses an image with the `-builder` suffix. It's designed to generate error pages and includes
> pre-rendered templates.

To set this up, we need two components:

* Nginx configuration file
* Dockerfile to build the image

Let's start with the Nginx configuration:

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

Now the Dockerfile:

```dockerfile
FROM docker.io/library/nginx:1.29-alpine

# override the default Nginx configuration
COPY --chown=nginx ./nginx.conf /etc/nginx/conf.d/default.conf

# copy prebuilt error pages from the "builder" image
# (instead of `ghost`, you can use any other template)
COPY --chown=nginx \
     --from=ghcr.io/tarampampam/error-pages:4-builder \
     /opt/html/ghost /usr/share/nginx/errorpages/_error-pages
```

Next, build and run the image:

```shell
docker build --tag your-nginx:local -f ./Dockerfile .
docker run --rm -p '8081:80/tcp' your-nginx:local
```

And voilà! Let's check that everything works as expected:

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
