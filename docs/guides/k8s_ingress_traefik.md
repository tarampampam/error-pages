# Use error-pages with Traefik in Kubernetes

> [!IMPORTANT]
> I am not a Kubernetes expert. This guide is **not a production-ready solution** - it is a working starting point
> and a cheat sheet for `error-pages` users. Everything described here has been tested and works, but every real
> deployment is different: validate the configuration, review the security implications, and adapt it to your own
> environment before using it in production.
>
> Contributions and improvements are very welcome - feel free to open a PR and I will happily accept it.

This guide wires `error-pages` as the global error handler for Traefik when used as a Kubernetes Ingress controller.
When any backend returns a status code in the `400–599` range, Traefik's
[errors middleware](https://doc.traefik.io/traefik/middlewares/http/errorpages/) intercepts the response and forwards
it to `error-pages`, which returns a styled page in the format requested by the client.

The middleware is attached at the **entrypoint level**: a single configuration change automatically covers every
route on the `web` entrypoint, with no per-Ingress annotation required. Traefik also supports two narrower opt-in
approaches - see the alternatives below before starting if you prefer per-service wiring.

## Alternative approaches

This guide uses the entrypoint approach - the closest equivalent to ingress-nginx's `custom-http-errors`, where a
single change applies to all backends at once. Traefik also supports two narrower alternatives that scope the
middleware to individual routes.

### Per-Ingress annotation

Attach the middleware to a single standard `Ingress` resource via annotation. Only that specific route gets styled
error pages; everything else remains unaffected:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: httpbin
  annotations:
    # format: {middleware-namespace}-{middleware-name}@kubernetescrd
    traefik.ingress.kubernetes.io/router.middlewares: "error-pages-error-pages@kubernetescrd"
spec:
  ingressClassName: traefik
  rules:
  - host: httpbin.localtest.me
    http:
      paths:
      - path: /
        pathType: Prefix
        backend: {service: {name: httpbin, port: {number: 80}}}
```

No `helm upgrade` for Traefik is required - Traefik picks up the annotation immediately. The trade-off is that every
Ingress that should use `error-pages` must explicitly include the annotation.

> [!NOTE]
> Both alternatives reference a `Middleware` from the `error-pages` namespace within resources that live in a
> different namespace. This requires `providers.kubernetesCRD.allowCrossNamespace=true` in Traefik - add
> `--set providers.kubernetesCRD.allowCrossNamespace=true` to the initial `helm upgrade --install traefik`
> command if you plan to use either approach.

See the [Kubernetes Ingress routing reference](https://doc.traefik.io/traefik/routing/providers/kubernetes-ingress/)
for the full list of supported annotations.

### IngressRoute CRD

Traefik's native `IngressRoute` CRD lets you reference a `Middleware` CRD directly in the route definition, without
annotations. This approach exposes more of Traefik's routing capabilities (priority, TCP/UDP routes, TLS options)
but requires authoring Traefik-specific CRDs instead of standard Kubernetes `Ingress` resources:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: httpbin
  namespace: default
spec:
  entryPoints: [web]
  routes:
  - match: Host(`httpbin.localtest.me`)
    kind: Rule
    middlewares:
    - name: error-pages
      namespace: error-pages # cross-namespace ref; requires allowCrossNamespace: true on Traefik
    services: [{name: httpbin, port: 80}]
```

The `Middleware` CRD created by the error-pages Helm chart (`traefikMiddleware.enabled=true`) works with both
approaches - no extra configuration of `error-pages` is required regardless of which routing style you choose.

See the [IngressRoute CRD reference](https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/) and
the [errors middleware reference](https://doc.traefik.io/traefik/middlewares/http/errorpages/) for full details.

## Local cluster setup

Follow the [kind guide](install_k8s_kind.md) to install the prerequisites.

## Install Traefik

Add the Helm repository and install Traefik as a DaemonSet with kind-compatible settings:

```shell
$ helm repo add traefik https://traefik.github.io/charts && helm repo update traefik
...Successfully got an update from the "traefik" chart repository
Update Complete. ⎈Happy Helming!⎈
```

> [!NOTE]
> Get the latest chart version from the [GitHub releases](https://github.com/traefik/traefik-helm-chart/releases/latest)
> page.

```shell
$ helm upgrade --install traefik traefik/traefik \
  --version 39.0.8 \
  --namespace traefik \
  --create-namespace \
  --set deployment.kind=DaemonSet \
  --set 'updateStrategy.rollingUpdate.maxSurge=0' \
  --set 'updateStrategy.rollingUpdate.maxUnavailable=1' \
  --set-string 'nodeSelector.ingress-ready=true' \
  --set 'tolerations[0].key=node-role.kubernetes.io/control-plane' \
  --set 'tolerations[0].operator=Exists' \
  --set 'tolerations[0].effect=NoSchedule' \
  --set 'ports.web.hostPort=80' \
  --set 'ports.websecure.hostPort=443' \
  --set service.type=NodePort \
  --wait \
  --timeout=90s

Release "traefik" does not exist. Installing it now.
NAME: traefik
LAST DEPLOYED: ...
NAMESPACE: traefik
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

A few kind-specific flags worth noting:

* `deployment.kind=DaemonSet` - runs exactly one Traefik pod per node; in a single-node kind cluster, this means one
  pod total. `hostPort` binding on a Deployment would also work, but a DaemonSet is the idiomatic choice when you want
  exactly one instance per node.
* `updateStrategy.rollingUpdate.maxSurge=0` / `maxUnavailable=1` - the Traefik chart defaults to `maxSurge=1`, which
  tries to start a new pod before stopping the old one. In kind, both pods would compete for the same host port, and
  the new pod would get stuck in `Pending`. Setting `maxSurge=0` ensures the old pod is stopped first.
* `ports.web.hostPort=80` - binds port 80 directly on the kind node. Combined with `extraPortMappings`, this makes
  Traefik reachable on `localhost:80`.

Verify the DaemonSet pods are running:

```shell
$ kubectl get pods --namespace traefik
NAME            READY   STATUS    RESTARTS   AGE
traefik-8zvhf   1/1     Running   0          84s
```

Verify Traefik is reachable on port 80. There are no Ingress rules yet, so it returns its built-in 404:

```shell
$ curl -so /dev/null -w "%{http_code}" http://localhost/
404

$ curl -s http://localhost/
404 page not found
```

## Deploy a test application

[go-httpbin](https://github.com/mccutchen/go-httpbin) is a small HTTP testing service; its `/status/{code}` endpoint
returns any requested status code - handy for testing error page interception.

Save the following to `httpbin.yaml`:

```yaml
# File: httpbin.yaml
apiVersion: apps/v1
kind: Deployment
metadata: {name: httpbin}
spec:
  replicas: 1
  selector: {matchLabels: {app: httpbin}}
  template:
    metadata: {labels: {app: httpbin}}
    spec: {containers: [{name: httpbin, image: ghcr.io/mccutchen/go-httpbin:2.22, ports: [{containerPort: 8080}]}]}
---
apiVersion: v1
kind: Service
metadata: {name: httpbin}
spec:
  selector: {app: httpbin}
  ports: [{port: 80, targetPort: 8080}]
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata: {name: httpbin}
spec:
  ingressClassName: traefik
  rules:
  - host: httpbin.localtest.me
    http: {paths: [{path: /, pathType: Prefix, backend: {service: {name: httpbin, port: {number: 80}}}}]}
```

`localtest.me` is a public DNS wildcard: `*.localtest.me` always resolves to `127.0.0.1`, so this works offline and
requires no `/etc/hosts` entry.

```shell
$ kubectl apply -f httpbin.yaml && kubectl rollout status deployment/httpbin --timeout=60s
deployment.apps/httpbin created
service/httpbin created
ingress.networking.k8s.io/httpbin created
deployment "httpbin" successfully rolled out
```

Verify the app responds normally:

```shell
$ curl -s http://httpbin.localtest.me/get
{
  "args": {},
  "headers": {
    "Accept": [
      "*/*"
    ],
    "Host": [
      "httpbin.localtest.me"
    ],
    "User-Agent": [
      "curl/8.11.1"
    ],
    "...": ["..."]
  },
  "method": "GET",
  "origin": "172.20.0.1",
  "url": "http://httpbin.localtest.me/get"
}
```

## Default error pages

Before `error-pages` is wired in, backend errors pass through as raw HTTP responses. go-httpbin's `/status/{code}`
returns the requested status code with an empty body - no message, no description:

```shell
$ curl -si http://httpbin.localtest.me/status/404
HTTP/1.1 404 Not Found
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
Content-Length: 0
Content-Type: text/plain; charset=utf-8
Date: ...
```

```shell
$ curl -si http://httpbin.localtest.me/status/503
HTTP/1.1 503 Service Unavailable
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
Content-Length: 0
Content-Type: text/plain; charset=utf-8
Date: ...
```

Requests that match no Ingress rule at all are served by Traefik's built-in 404 page:

```shell
$ curl -s http://unknown.localtest.me/ | head -n 5
404 page not found
```

## 🔥 Install error-pages

Install error-pages with `traefikMiddleware.enabled=true` - this creates a `Middleware` CRD in the `error-pages`
namespace that Traefik will use to intercept error responses:

> [!NOTE]
> Get the latest chart version from [ArtifactHub](https://artifacthub.io/packages/helm/error-pages/error-pages)
> or the [GitHub releases](https://github.com/tarampampam/error-pages/releases) page.

```shell
$ helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version X.Y.Z \
  --namespace error-pages \
  --create-namespace \
  --set config.htmlTemplate.name=l7 \
  --set traefikMiddleware.enabled=true \
  --wait \
  --timeout=60s

NAME: error-pages
LAST DEPLOYED: ...
NAMESPACE: error-pages
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

Verify the `Middleware` CRD was created:

```shell
$ kubectl get middleware --namespace error-pages
NAME          AGE
error-pages   10s
```

> [!NOTE]
> Unlike ingress-nginx, `config.sendSameHttpCode=true` is **not** required here. Traefik's errors middleware
> preserves the original backend status code and only replaces the response body - it does not use the status
> code returned by the error service.

## Wire Traefik to error-pages

Attach the `Middleware` CRD to Traefik's `web` entrypoint as a static argument. Every route using this entrypoint
will then automatically have its 4xx/5xx responses intercepted and styled:

```shell
$ helm upgrade traefik traefik/traefik \
  --version 39.0.8 \
  --namespace traefik \
  --reuse-values \
  --set 'additionalArguments[0]=--entrypoints.web.http.middlewares=error-pages-error-pages@kubernetescrd' \
  --wait \
  --timeout=60s

Release "traefik" has been upgraded. Happy Helming!
NAME: traefik
LAST DEPLOYED: ...
NAMESPACE: traefik
STATUS: deployed
REVISION: 2
DESCRIPTION: Upgrade complete
```

The middleware identifier format is `{namespace}-{name}@kubernetescrd`: namespace `error-pages`, name
`error-pages`, provider `kubernetescrd`.

## Verify custom error pages

The same requests that previously returned empty responses now return styled error pages:

```shell
$ curl -s -H "Accept: text/html" http://httpbin.localtest.me/status/404 | head -n 15
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
$ curl -s -H "Accept: application/json" http://httpbin.localtest.me/status/404
{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}
```

```shell
$ curl -s -H "Accept: application/xml" http://httpbin.localtest.me/status/503
<?xml version="1.0" encoding="utf-8"?>
<error>
  <code>503</code>
  <message>Service Unavailable</message>
  <description>The server is temporarily overloading or down</description>
</error>
```

The middleware applies to every **matched** backend on the `web` entrypoint - any service deployed with
`ingressClassName: traefik` gets styled error pages with no additional configuration.

Requests that do not match any Ingress rule are handled differently. The errors middleware only runs when a matched
route calls a backend that returns 4xx/5xx. When no route matches, Traefik handles the request internally and
returns its own plain-text response - the middleware is never part of the processing chain:

```shell
$ curl -s http://unknown.localtest.me/
404 page not found
```

To cover unmatched hostnames as well, add a catch-all `IngressRoute` that forwards any request not claimed by a
specific Ingress directly to `error-pages` at very low priority:

```yaml
# File: catch-all.yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: error-pages-catch-all
  namespace: error-pages
spec:
  entryPoints: [web]
  routes:
  - match: HostRegexp(`.+`)
    kind: Rule
    priority: 1
    services: [{name: error-pages, port: 8080}]
```

```shell
$ kubectl apply -f catch-all.yaml
ingressroute.traefik.io/error-pages-catch-all created
```

`HostRegexp(.+)` matches any hostname. The explicit `priority: 1` ensures that specific Ingress rules always take
precedence - this route is only used when nothing else matches. Because the `IngressRoute` and the `error-pages`
service are in the same namespace, no `allowCrossNamespace` setting is required.

The catch-all routes requests directly to `error-pages` as a backend rather than through the errors middleware, so
set `config.sendSameHttpCode=true` on the error-pages release to return the correct HTTP status code instead of
`200 OK`:

```shell
$ helm upgrade error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version X.Y.Z \
  --namespace error-pages \
  --reuse-values \
  --set config.sendSameHttpCode=true
```

Unmatched hostnames now return a styled error page instead of Traefik's default fallback:

```shell
$ curl -s -H "Accept: application/json" http://unknown.localtest.me/
{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}
```

```shell
$ curl -so /dev/null -w "%{http_code}" http://unknown.localtest.me/
404
```

## Cleanup

Delete the kind cluster to remove all resources at once:

```shell
kind delete cluster --name error-pages-test-cluster
```
