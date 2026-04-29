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
  #query: "/{status}.html"
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
  addCode: |
    499=Client Closed Request|The client closed the connection before the server finished responding.
    4**=Client Error|Something went wrong on the client side.
```

{{ template "chart.valuesSection" . }}
