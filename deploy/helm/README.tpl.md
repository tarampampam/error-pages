# {{ template "chart.name" . }}

{{ template "chart.description" . }}

## Installing the chart

```shell
# Install a specific version
helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version {{ template "chart.version" . }}

# Install with custom values file
helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version {{ template "chart.version" . }} \
  --values my-values.yaml
```

## Upgrading

```shell
helm upgrade error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages
```

## Use cases

### ingress-nginx - default backend

Route all unhandled error responses through error-pages when using [ingress-nginx](https://kubernetes.github.io/ingress-nginx/).

`values.yaml`:

```yaml
config:
  sendSameHttpCode: true
```

For more details, please, refer to the [project's documentation](https://github.com/tarampampam/error-pages#readme).

### Traefik - errors middleware

Use the built-in `traefikMiddleware` to let [Traefik](https://doc.traefik.io/traefik/) intercept error responses
and forward them to error-pages.

`values.yaml`:

```yaml
traefikMiddleware:
  enabled: true # creates a Middleware CRD in the same namespace
  #statusCodes: ["400-599"]
  #query: "/{status}"
```

For more details, please, refer to the [project's documentation](https://github.com/tarampampam/error-pages#readme).

### Template rotation

Serve a different built-in HTML template on each request (great for a bit of personality in staging environments):

```yaml
config:
  htmlTemplate:
    rotationMode: random-on-each-request
```

### Custom HTML template

Load a custom Go template from a ConfigMap volume, a URL, or inline text:

```yaml
# Inline template (small templates only)
config:
  htmlTemplate:
    custom: |
      <!DOCTYPE html>
      <html>
        <body><h1>{{ "{{" }} .StatusCode {{ "}}" }} - {{ "{{" }} .Message {{ "}}" }}</h1></body>
      </html>
```

```yaml
# From a URL fetched once at startup
config:
  htmlTemplate:
    custom: "https://example.com/my-error-template.html"
```

```yaml
# From a file mounted into the pod
deployment:
  volumes:
    - name: templates
      configMap:
        name: my-error-templates
  volumeMounts:
    - name: templates
      mountPath: /templates
      readOnly: true

config:
  htmlTemplate:
    custom: "/templates/error.html"
```

### Custom HTTP status codes

Override descriptions or add non-standard codes (e.g. `499`, `4**` wildcard):

```yaml
config:
  addCode:
    - {code: "4**", message: "Client Error", description: "Something went wrong on the client side"}
    - code: "499"
      message: "Client Closed Request"
      description: "The client closed the connection before the server finished responding"
```

Via `--set` (`--set-string` is required for numeric-looking codes like `499`):

```shell
helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --set-string 'config.addCode[0].code=4**' \
  --set 'config.addCode[0].message=Client Error' \
  --set-string 'config.addCode[1].code=499' \
  --set 'config.addCode[1].message=Client Closed Request' \
  --set 'config.addCode[1].description=The client closed the connection before the server finished responding'
```

### Adding extra links

Display additional links (status page, contact, privacy policy, etc.) on all error pages:

```yaml
config:
  addLink:
    - {label: "Status Page", url: "https://status.example.com"}
    - {label: "Contact Support", url: "https://example.com/contact"}
    - {label: "Privacy Policy", url: "https://example.com/privacy"}
```

Via `--set`:

```shell
helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --set 'config.addLink[0].label=Status Page' \
  --set 'config.addLink[0].url=https://status.example.com' \
  --set 'config.addLink[1].label=Contact Support' \
  --set 'config.addLink[1].url=https://example.com/contact'
```

## 💊 Support

If you need a chart option that doesn't exist yet, or something isn't working as expected, please
[open an issue](https://github.com/tarampampam/error-pages/issues/new/choose) - I'll be happy to help.

{{ template "chart.valuesSection" . }}
