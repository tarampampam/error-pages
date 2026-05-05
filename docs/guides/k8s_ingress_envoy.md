# Use error-pages with Envoy Gateway in Kubernetes

> [!IMPORTANT]
> I am not a Kubernetes expert. This guide is **not a production-ready solution** - it is a working starting point
> and a cheat sheet for `error-pages` users. Everything described here has been tested and works, but every real
> deployment is different: validate the configuration, review the security implications, and adapt it to your own
> environment before using it in production.
>
> Contributions and improvements are very welcome - feel free to open a PR and I will happily accept it.

This guide wires `error-pages` as a centralized error handler for Envoy Gateway in two ways:

1. **Route-miss 404s** - catch-all HTTPRoute for requests that don't match any specific route
2. **Backend errors (4xx/5xx)** - `BackendTrafficPolicy` that internally redirects error responses to error-pages

## Local cluster setup

Follow the [kind guide](install_k8s_kind.md) to install the prerequisites.

## Install Envoy Gateway

> [!NOTE]
> Get the latest chart version from the [GitHub releases](https://github.com/envoyproxy/gateway/releases/latest) page.

```shell
$ helm upgrade --install envoy-gateway oci://docker.io/envoyproxy/gateway-helm \
  --version 1.7.2 \
  --namespace envoy-gateway-system \
  --create-namespace \
  --wait \
  --timeout=90s

NAME: envoy-gateway
LAST DEPLOYED: ...
NAMESPACE: envoy-gateway-system
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

## Configure Envoy proxy for kind

> [!NOTE]
> KIND-specific. In a cloud environment with a real LoadBalancer - skip this.

Envoy Gateway provisions a LoadBalancer service that stays `<pending>` in kind. To expose port 80 on the host,
configure `hostPort` via an `EnvoyProxy` resource. Envoy internally shifts privileged port 80 to containerPort
10080 (+10000 offset), so `hostPort: 80` must be mapped to `containerPort: 10080`.

```yaml
# File: envoy-proxy-config.yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata: {name: kind-proxy-config, namespace: envoy-gateway-system}
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  tolerations: [{key: "node-role.kubernetes.io/control-plane", operator: Exists, effect: NoSchedule}]
                  nodeSelector: {ingress-ready: "true"}
                  containers: [{name: envoy, ports: [{containerPort: 10080, hostPort: 80, protocol: TCP}]}]
```

```shell
$ kubectl apply -f envoy-proxy-config.yaml
envoyproxy.gateway.envoyproxy.io/kind-proxy-config created
```

## Create a GatewayClass

Create a `GatewayClass` that references the `EnvoyProxy` configuration above:

```yaml
# File: gateway-class.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata: {name: envoy-gateway}
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: kind-proxy-config
    namespace: envoy-gateway-system
```

```shell
$ kubectl apply -f gateway-class.yaml
gatewayclass.gateway.networking.k8s.io/envoy-gateway created
```

## Create a Gateway

```yaml
# File: gateway.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata: {name: envoy-gateway, namespace: default}
spec:
  gatewayClassName: envoy-gateway
  listeners: [{name: http, port: 80, protocol: HTTP}]
```

```shell
$ kubectl apply -f gateway.yaml
gateway.gateway.networking.k8s.io/envoy-gateway created
```

```shell
$ kubectl rollout status deployment \
  -l "gateway.envoyproxy.io/owning-gateway-name=envoy-gateway" \
  -n envoy-gateway-system \
  --timeout=90s
deployment "envoy-default-envoy-gateway-12b6bb46" successfully rolled out
```

> [!NOTE]
> In KIND, `kubectl get gateway envoy-gateway` shows `PROGRAMMED=False` with no `ADDRESS` - the LoadBalancer
> service never gets an external IP. This is cosmetic; the gateway works via `hostPort`. Verify with
> `curl -si http://localhost/` (expect an empty-body 404).

## Deploy a test application

[go-httpbin](https://github.com/mccutchen/go-httpbin) returns any requested status code via `/status/{code}`,
making it perfect for testing error page interception.

`localtest.me` is a public DNS wildcard resolving to `127.0.0.1` - no `/etc/hosts` entry needed.

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
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata: {name: httpbin}
spec:
  parentRefs: [{name: envoy-gateway, namespace: default}]
  hostnames: [httpbin.localtest.me]
  rules:
    - matches: [{path: {type: PathPrefix, value: /}}]
      backendRefs: [{name: httpbin, port: 80}]
```

```shell
$ kubectl apply -f httpbin.yaml && kubectl rollout status deployment/httpbin --timeout=60s
deployment.apps/httpbin created
service/httpbin created
httproute.gateway.networking.k8s.io/httpbin created
deployment "httpbin" successfully rolled out
```

**Before** `error-pages` is wired in, both backend errors and route-miss responses have an empty body:

```shell
$ curl -si http://httpbin.localtest.me/status/404
HTTP/1.1 404 Not Found
access-control-allow-credentials: true
access-control-allow-origin: *
content-type: text/plain; charset=utf-8
date: ...
content-length: 0

$ curl -si http://unknown.localtest.me/
HTTP/1.1 404 Not Found
date: ...
content-length: 0
```

## 🔥 Install error-pages

`config.sendSameHttpCode=true` is required - it makes error-pages return the same HTTP status code as the requested
path (e.g., `/404` → HTTP 404, `/503` → HTTP 503).

> [!NOTE]
> Get the latest chart version from [ArtifactHub](https://artifacthub.io/packages/helm/error-pages/error-pages)
> or the [GitHub releases](https://github.com/tarampampam/error-pages/releases) page.

```shell
$ helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version X.Y.Z \
  --namespace error-pages \
  --create-namespace \
  --set config.sendSameHttpCode=true \
  --set config.htmlTemplate.name=l7 \
  --wait \
  --timeout=60s

NAME: error-pages
LAST DEPLOYED: ...
NAMESPACE: error-pages
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

## Wire Envoy Gateway to error-pages

### Step 1: ReferenceGrant for cross-namespace access

HTTPRoutes in `default` reference the `error-pages` Service in the `error-pages` namespace. A `ReferenceGrant`
in the target namespace is required, or routes fail with `RefNotPermitted`.

```yaml
# File: reference-grant.yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata: {name: allow-default-to-error-pages, namespace: error-pages}
spec:
  from: [{group: gateway.networking.k8s.io, kind: HTTPRoute, namespace: default}]
  to: [{group: "", kind: Service, name: error-pages}]
```

```shell
$ kubectl apply -f reference-grant.yaml
referencegrant.gateway.networking.k8s.io/allow-default-to-error-pages created
```

### Step 2: HTTPRoute for the error-pages service

The `BackendTrafficPolicy` redirect in Step 4 works by making Envoy re-route requests to a URL that must exist in
Envoy's routing table. Without this HTTPRoute, the redirect hostname has no route and the redirect silently falls back
to the original backend response.

```yaml
# File: error-pages-route.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata: {name: error-pages-direct, namespace: default}
spec:
  parentRefs: [{name: envoy-gateway, namespace: default}]
  hostnames: [error-pages.localtest.me]
  rules: [{backendRefs: [{name: error-pages, namespace: error-pages, port: 8080}]}]
```

```shell
$ kubectl apply -f error-pages-route.yaml
httproute.gateway.networking.k8s.io/error-pages-direct created
```

### Step 3: Catch-all HTTPRoute for route-miss 404

> [!IMPORTANT]
> Must be in the same namespace as the Gateway (`default`) - Envoy Gateway's default `allowedRoutes.namespaces.from: Same`
> only allows routes from the same namespace. Omitting `hostnames` makes it match any request not claimed by a
> more-specific route.

```yaml
# File: error-pages-catchall.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata: {name: error-pages-catchall, namespace: default}
spec:
  parentRefs: [{name: envoy-gateway, namespace: default}]
  rules:
    - matches: [{path: {type: PathPrefix, value: /}}]
      backendRefs: [{name: error-pages, namespace: error-pages, port: 8080}]
```

```shell
$ kubectl apply -f error-pages-catchall.yaml
httproute.gateway.networking.k8s.io/error-pages-catchall created
```

### Step 4: BackendTrafficPolicy for backend error interception

> [!WARNING]
> **Envoy Gateway does not support dynamic status code substitution in `replaceFullPath`**. There is no `{status}`,
> `%RESPONSE_CODE%`, or any other placeholder - the value is a plain static string baked into the xDS config at
> translation time (at least for now - May 2026, Envoy Gateway chart v1.7.2).
>
> **Each intercepted status code requires its own explicit rule**. Catch-all ranges like `400-499 → /404`
> silently map every distinct backend error (403, 429, 422, etc.) to a single misleading page and report the
> wrong status code to the client.

```yaml
# File: error-pages-policy.yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata: {name: error-pages-policy, namespace: default}
spec:
  targetRefs: [{group: gateway.networking.k8s.io, kind: HTTPRoute, name: httpbin}]
  responseOverride:
    # each status code needs its own rule - Envoy Gateway has no dynamic path substitution
    - match: {statusCodes: [{type: Value, value: 404}]}
      redirect:
        scheme: http
        hostname: error-pages.localtest.me
        port: 80
        path: {type: ReplaceFullPath, replaceFullPath: /404}
    - match:
        statusCodes: [{type: Value, value: 500}]
      redirect:
        scheme: http
        hostname: error-pages.localtest.me
        port: 80
        path: {type: ReplaceFullPath, replaceFullPath: /500}
    - match: {statusCodes: [{type: Value, value: 503}]}
      redirect:
        scheme: http
        hostname: error-pages.localtest.me
        port: 80
        path: {type: ReplaceFullPath, replaceFullPath: /503}
    # add rules for every code your backends can return: 400, 401, 403, 429, 502, 504, etc.
```

```shell
$ kubectl apply -f error-pages-policy.yaml
backendtrafficpolicy.gateway.envoyproxy.io/error-pages-policy created
```

The `redirect` performs an **internal** proxy - Envoy fetches the error-pages response and returns it to the client
transparently. The client receives the status code that `error-pages` returns for the path (e.g., `/404` → HTTP 404).
Backend errors with no matching rule pass through to the client unchanged.

## Limitations

- **All HTTPRoutes must be in the Gateway namespace**. Envoy Gateway's default `allowedRoutes.namespaces.from: Same`
  requires all routes in the same namespace as the Gateway (`default`). Cross-namespace `backendRef` to the
  `error-pages` Service is allowed via the `ReferenceGrant` from Step 1.
- **Catch-all priority**. Application routes (specific hostnames) always take precedence over the catch-all.

## Verify custom error pages

Backend error - styled HTML:

```shell
$ curl -si -H "Accept: text/html" http://httpbin.localtest.me/status/404 | head -n 13
HTTP/1.1 404 Not Found
content-length: 53355
content-type: text/html; charset=utf-8
x-request-id: d8d3e89c-206b-4afa-bb47-d141fa6a6afc
x-robots-tag: noindex, nofollow, nosnippet, noarchive
date: ...

<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="robots" content="nofollow,noarchive,noindex">
  <title data-l10n>Not Found</title>
```

Backend error - JSON via Accept header:

```shell
$ curl -si -H "Accept: application/json" http://httpbin.localtest.me/status/404
HTTP/1.1 404 Not Found
content-length: 124
content-type: application/json; charset=utf-8
x-request-id: c34021bb-5464-44db-80ec-e39e87a9c04c
x-robots-tag: noindex, nofollow, nosnippet, noarchive
date: ...

{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}
```

Backend error - XML:

```shell
$ curl -si -H "X-Format: application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" http://httpbin.localtest.me/status/503
HTTP/1.1 503 Service Unavailable
content-length: 191
content-type: application/xml; charset=utf-8
retry-after: 120
x-request-id: 343fdfab-c865-445c-b9fe-8fddcd2d71f3
x-robots-tag: noindex, nofollow, nosnippet, noarchive
date: ...

<?xml version="1.0" encoding="utf-8"?>
<error>
  <code>503</code>
  <message>Service Unavailable</message>
  <description>The server is temporarily overloading or down</description>
</error>
```

Route-miss - served by catch-all, also styled:

```shell
$ curl -si -H "Accept: application/json" http://unknown.localtest.me/
HTTP/1.1 404 Not Found
content-length: 124
content-type: application/json; charset=utf-8
x-request-id: 6fe52817-439c-40bf-ae00-7141892de7c9
x-robots-tag: noindex, nofollow, nosnippet, noarchive
date: ...

{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}
```

Status codes without a rule pass through unchanged:

```shell
$ curl -si http://httpbin.localtest.me/status/403
HTTP/1.1 403 Forbidden
access-control-allow-credentials: true
access-control-allow-origin: *
content-type: text/plain; charset=utf-8
date: ...
content-length: 0
```

### Cleanup

Delete the kind cluster to remove all resources at once:

```shell
kind delete cluster --name error-pages-test-cluster
```
