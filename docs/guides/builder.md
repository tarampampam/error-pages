# Generate error pages using built-in or custom templates

> [!IMPORTANT]
> Here, we use an image with the `-builder` suffix. It's designed for generating error pages and includes
> pre-rendered templates.

Generating a set of error pages is straightforward. If you prefer to use your own template, start by creating one.
For example:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <title>{{ .StatusCode }}</title>
</head>
<body>
  <h1>{{ .Message }}: {{ .Description }}</h1>
</body>
</html>
```

Save it as `my-super-theme.html`. Then generate your error pages with the following command:

```shell
mkdir ./out
docker run --rm \
  -v "$(pwd)/my-super-theme.html:/opt/my-template.html:ro" \
  -v "$(pwd)/out:/opt/out:rw" \
  -u $(id -u):$(id -g) \
  ghcr.io/tarampampam/error-pages:4-builder \
  --template /opt/my-template.html \
  --out /opt/out \
  --index
```

This will generate error pages based on your template in the specified output directory:

```shell
tree ./out/

├── 400.html
├── 401.html
├── 403.html
├── 404.html
├── 405.html
├── 407.html
├── 408.html
├── 409.html
├── 410.html
├── 411.html
├── 412.html
├── 413.html
├── 416.html
├── 418.html
├── 429.html
├── 500.html
├── 502.html
├── 503.html
├── 504.html
├── 505.html
└── index.html
```

```shell
cat ./out/403.html
<!DOCTYPE html>
<html lang="en">
<head>
  <title>403</title>
</head>
<body>
  <h1>Forbidden: Access is forbidden to the requested page</h1>
</body>
</html>
```
