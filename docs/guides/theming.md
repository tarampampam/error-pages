# Start the HTTP server with a custom template (theme)

First, create your own template file - for example, `my-super-theme.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <title>{{ .StatusCode }}</title>
</head>
<body>
  <h1>YEAH! {{ .Message }}: {{ .Description }}</h1>
</body>
</html>
```

Then start the server:

```shell
docker run --rm \
  -v "$(pwd)/my-super-theme.html:/opt/my-template.html:ro" \
  -p '8080:8080/tcp' ghcr.io/tarampampam/error-pages:4 \
  --html-template /opt/my-template.html
```

Test it:

```shell
curl -H "Accept: text/html" http://127.0.0.1:8080/503

<!DOCTYPE html>
<html lang="en">
<head>
  <title>503</title>
</head>
<body>
  <h1>YEAH! Service Unavailable: The server is temporarily overloading or down</h1>
</body>
</html>
```

## Using docker-compose

Now let's do the same thing with `docker-compose` and environment variables:

```yaml
# yaml-language-server: $schema=https://cdn.jsdelivr.net/gh/compose-spec/compose-spec@master/schema/compose-spec.json

services:
  error-pages:
    image: ghcr.io/tarampampam/error-pages:4
    environment:
      #language=html
      HTML_TEMPLATE: |
        <!DOCTYPE html>
        <html lang="en">
        <head>
          <title>{{ .StatusCode }}</title>
        </head>
        <body>
          <h1>{{ .Description }}</h1>
        </body>
        </html>
      #language=json
      JSON_TEMPLATE: |
        {
          "mission": "successfully failed",
          "message": {{ .Message | toJson }}
        }
    ports:
      - '8080:8080/tcp'
```

Save it as `compose.yml`, then run:

```shell
docker compose -f compose.yml up
```

Run a few tests:

```shell
curl -H "Accept: text/html" http://127.0.0.1:8080/503

<!DOCTYPE html>
<html lang="en">
<head>
  <title>503</title>
</head>
<body>
  <h1>The server is temporarily overloading or down</h1>
</body>
</html>
```

```shell
curl http://127.0.0.1:8080/505.json

{
  "mission": "successfully failed",
  "message": "HTTP Version Not Supported"
}
```
