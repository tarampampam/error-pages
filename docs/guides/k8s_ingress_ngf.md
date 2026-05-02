# Use error pages with NGINX Gateway Fabric in Kubernetes

> [!IMPORTANT]
> I am not a Kubernetes expert. This guide is **not a production-ready solution** - it is a working starting point
> and a cheat sheet for `error-pages` users. Everything described here has been tested and works, but every real
> deployment is different: validate the configuration, review the security implications, and adapt it to your own
> environment before using it in production.
>
> Contributions and improvements are very welcome - feel free to open a PR and I will happily accept it.

This guide wires `error-pages` as the custom error backend for
[NGINX Gateway Fabric](https://github.com/nginx/nginx-gateway-fabric) (NGF). When any backend routed through the
Gateway returns an error status code, an NGINX `SnippetsPolicy` intercepts the response and proxies it to
`error-pages`, which returns a styled page in the format the client requested.

> [!NOTE]
> Unlike ingress-nginx (which has a built-in default backend concept), NGF uses the
> [SnippetsPolicy](https://docs.nginx.com/nginx-gateway-fabric/reference/api/#gateway.nginx.org/v1alpha1.SnippetsPolicy)
> API to inject raw NGINX configuration directives. `SnippetsPolicy` is an **alpha** feature and must be explicitly
> enabled at install time.
>
> `SnippetsPolicy` is the only available method because neither the Gateway API specification nor NGF provide a
> first-class API for custom error pages:
>
> - [kubernetes-sigs/gateway-api#1998](https://github.com/kubernetes-sigs/gateway-api/issues/1998)
> - [kubernetes-sigs/gateway-api#2826](https://github.com/kubernetes-sigs/gateway-api/issues/2826)
> - [nginx/nginx-gateway-fabric#4967](https://github.com/nginx/nginx-gateway-fabric/issues/4967)
> - The [NGF Snippets documentation](https://docs.nginx.com/nginx-gateway-fabric/traffic-management/snippets/)
>   explicitly states they should be used "only in cases where Gateway API resources or NGINX extension policies
>   don't apply" - custom error pages are exactly that case.

## Local cluster setup

Follow the [kind guide](install_k8s_kind.md) to install the prerequisites.

## Install NGINX Gateway Fabric

### Install Gateway API CRDs

NGF implements the [Gateway API](https://gateway-api.sigs.k8s.io/) standard. Install the Gateway API CRDs first:

```shell
$ kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml
customresourcedefinition.apiextensions.k8s.io/gatewayclasses.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/gateways.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/grpcroutes.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/httproutes.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/referencegrants.gateway.networking.k8s.io created
```

### Install the NGF controller

> [!NOTE]
> Get the latest chart version from the [GitHub releases](https://github.com/nginx/nginx-gateway-fabric/releases/latest)
> page.

```shell
$ helm install ngf oci://ghcr.io/nginx/charts/nginx-gateway-fabric \
  --version 2.5.1 \
  --namespace nginx-gateway \
  --create-namespace \
  --set nginxGateway.snippets.enable=true \
  --set nginx.service.type=NodePort \
  --set 'nginx.container.hostPorts[0].port=80' \
  --set 'nginx.container.hostPorts[0].containerPort=80' \
  --set 'nginx.pod.tolerations[0].key=node-role.kubernetes.io/master' \
  --set 'nginx.pod.tolerations[0].effect=NoSchedule' \
  --set 'nginx.pod.tolerations[0].operator=Exists' \
  --set 'nginx.pod.tolerations[1].key=node-role.kubernetes.io/control-plane' \
  --set 'nginx.pod.tolerations[1].effect=NoSchedule' \
  --set 'nginx.pod.tolerations[1].operator=Exists' \
  --set-string 'nginx.pod.nodeSelector.ingress-ready=true' \
  --wait \
  --timeout=90s

NAME: ngf
LAST DEPLOYED: ...
NAMESPACE: nginx-gateway
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

Kind-specific settings explained:

- `nginxGateway.snippets.enable=true` - enables the `SnippetsPolicy` and `SnippetsFilter` alpha APIs, which are
  disabled by default. **This is required to inject the error interception NGINX config**.
- `nginx.service.type=NodePort` - avoids a `<pending>` external IP on the NGINX data plane Service. Traffic reaches
  NGINX via `hostPort` → `extraPortMappings`, not through the Service.
- `nginx.container.hostPorts` - binds port 80 of the NGINX data plane pod directly to the kind node's host port,
  so that traffic arriving at `localhost:80` reaches NGINX.
- `nginx.pod.tolerations` and `nginx.pod.nodeSelector` - allow the NGINX data plane pod to be scheduled on kind's
  control-plane node, which carries both `node-role.kubernetes.io/master` and
  `node-role.kubernetes.io/control-plane` taints. The `ingress-ready=true` label is set on the node by the kind
  cluster config.

Verify the controller pod is ready:

```shell
$ kubectl get pods --namespace nginx-gateway
NAME                                        READY   STATUS    RESTARTS   AGE
ngf-nginx-gateway-fabric-7d95c89b7b-xxxxx   1/1     Running   0          60s
```

### Create a Gateway

NGF uses a split control-plane / data-plane architecture. The control plane (`ngf-nginx-gateway-fabric`) is
installed at Helm install time, but the NGINX data plane pod is provisioned on demand when you create a `Gateway`
resource. Save the following to `gateway.yaml`:

```yaml
# File: gateway.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: nginx
  namespace: default
spec:
  gatewayClassName: nginx
  listeners:
    - name: http
      port: 80
      protocol: HTTP
```

```shell
$ kubectl apply -f gateway.yaml
gateway.gateway.networking.k8s.io/nginx created
```

Wait for the NGINX data plane pod to appear and become ready:

```shell
$ kubectl rollout status deployment/nginx-nginx --namespace default --timeout=90s
deployment "nginx-nginx" successfully rolled out

$ kubectl get gateway nginx
NAME    CLASS   ADDRESS        PROGRAMMED   AGE
nginx   nginx   10.96.96.113   True         30s
```

Verify NGF is reachable on port `80`. There are no HTTPRoutes yet, so NGINX returns its built-in 404:

```shell
$ curl -s http://localhost/ | head -n 5
<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx</center>
```

## Deploy a test application

[go-httpbin](https://github.com/mccutchen/go-httpbin) is a small HTTP testing service. Its `/status/{code}` endpoint
returns any requested status code, which makes it handy for testing error page interception.

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
  ports: [{port: 8080, targetPort: 8080}]
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata: {name: httpbin}
spec:
  parentRefs: [{name: nginx, namespace: default}]
  hostnames: [httpbin.localtest.me]
  rules:
    - matches: [{path: {type: PathPrefix, value: /}}]
      backendRefs: [{name: httpbin, port: 8080}]
```

`localtest.me` is a public DNS wildcard: `*.localtest.me` always resolves to `127.0.0.1`, so this works offline and
doesn't require an `/etc/hosts` entry.

```shell
$ kubectl apply -f httpbin.yaml && kubectl rollout status deployment/httpbin --timeout=60s
deployment.apps/httpbin created
service/httpbin created
httproute.gateway.networking.k8s.io/httpbin created
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

Before `error-pages` is wired in, backend errors pass through to the client as-is. Without `proxy_intercept_errors on`,
NGINX proxies the upstream response without modifying it. go-httpbin's `/status/{code}` returns the requested status
code with an empty body - no message, no description, nothing useful for an end user:

```shell
$ curl -si http://httpbin.localtest.me/status/404
HTTP/1.1 404 Not Found
Server: nginx
Date: ...
Content-Type: text/plain; charset=utf-8
Content-Length: 0
Connection: keep-alive
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
```

```shell
$ curl -si http://httpbin.localtest.me/status/503
HTTP/1.1 503 Service Unavailable
Server: nginx
Date: ...
Content-Type: text/plain; charset=utf-8
Content-Length: 0
Connection: keep-alive
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
```

## 🔥 Install error-pages

Setting **`config.sendSameHttpCode=true`** is critical for this integration - `error-pages` must return the same HTTP
status code as the error it renders, not `200`. Otherwise, NGINX will pass the `200` response to the client and the
actual error code will be lost.

> [!NOTE]
> Get the latest chart version from [ArtifactHub](https://artifacthub.io/packages/helm/error-pages/error-pages)
> or the [GitHub releases](https://github.com/tarampampam/error-pages/releases) page.

```shell
$ helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version X.Y.Z \
  --namespace error-pages \
  --create-namespace \
  --set config.sendSameHttpCode=true \
  --wait \
  --timeout=60s

NAME: error-pages
LAST DEPLOYED: ...
NAMESPACE: error-pages
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

## Wire NGINX Gateway Fabric to error-pages

NGF does not have a built-in default backend concept. Instead, use a `SnippetsPolicy` to inject raw NGINX configuration
into every server block managed by the Gateway. The snippet enables `proxy_intercept_errors` and defines a named
location that proxies intercepted errors to `error-pages`.

Save the following to `error-pages-policy.yaml`:

```yaml
# File: error-pages-policy.yaml
apiVersion: gateway.nginx.org/v1alpha1
kind: SnippetsPolicy
metadata:
  name: error-pages
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: nginx
  snippets:
    - context: http.server
      value: |
        proxy_intercept_errors on;
        error_page 400 401 403 404 405 408 409 410 429 500 502 503 504 = @error_pages;
        location @error_pages {
          internal;
          proxy_intercept_errors off;
          proxy_pass http://error-pages.error-pages.svc.cluster.local:8080;
          proxy_set_header X-Code $upstream_status;
          proxy_set_header Accept $http_accept;
          proxy_set_header X-Original-URI $request_uri;
          proxy_set_header Host $host;
        }
```

```shell
$ kubectl apply -f error-pages-policy.yaml
snippetspolicy.gateway.nginx.org/error-pages created
```

Wait for the policy to be accepted:

```shell
$ kubectl get snippetspolicy error-pages
NAME          AGE
error-pages   5s

$ kubectl get snippetspolicy error-pages -o jsonpath='{.status.ancestors[0].conditions[0].message}'
The Policy is accepted
```

A few things in the snippet are worth explaining:

* **`proxy_intercept_errors on`** - tells NGINX to intercept upstream responses with error status codes instead of
  passing them directly to the client. Without this directive, the raw backend error passes through.
* **`error_page ... = @error_pages`** - lists the status codes to intercept and redirects them to the `@error_pages`
  named location. The `=` preserves the status code from the error-pages response.
* **`proxy_intercept_errors off`** inside `@error_pages` - prevents NGINX from intercepting the response from
  `error-pages` itself. Without this, NGINX would try to intercept `error-pages`'s own 4xx/5xx response and loop.
* **`proxy_set_header X-Code $upstream_status`** - passes the original error code to `error-pages`. Note that
  `$upstream_status` (not `$status`) must be used here: in the context of a named location triggered by `error_page`,
  `$status` is `0`, whereas `$upstream_status` retains the error code from the backend that triggered the interception.

## Verify custom error pages

The same requests that previously returned bare NGINX error pages now return styled `error-pages` responses.

Plain text (default, curl-friendly):

```shell
$ curl -s http://httpbin.localtest.me/status/404
Error 404: Not Found
The server can not find the requested page
```

JSON (via `Accept` header):

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
$ curl -s -H "Accept: application/json" http://httpbin.localtest.me/status/500
{
  "error": true,
  "code": 500,
  "message": "Internal Server Error",
  "description": "The server met an unexpected condition"
}
```

XML (via `Accept` header):

```shell
$ curl -s -H "Accept: application/xml" http://httpbin.localtest.me/status/503
<?xml version="1.0" encoding="utf-8"?>
<error>
  <code>503</code>
  <message>Service Unavailable</message>
  <description>The server is temporarily overloading or down</description>
</error>
```

The HTTP status code of each response matches the intercepted error:

```shell
$ curl -s -o /dev/null -w "%{http_code}" http://httpbin.localtest.me/status/500
500
```

### Handle requests to unmatched hostnames (optional)

By default, requests to hostnames not covered by any HTTPRoute are handled by NGF's auto-generated `default_server`
block, which does `return 404` without any snippet - those responses pass through as plain NGINX HTML.

To serve styled error pages for those requests too, add a catch-all HTTPRoute with no `hostnames` field. Gateway API
routes without a `hostnames` field match any hostname in the listener. NGF creates a dedicated NGINX server block
(`server_name ~^;`) for such a route, and `SnippetsPolicy` - which targets the whole Gateway - injects the snippet
into it just like any other server block.

Because `error-pages` lives in the `error-pages` namespace and the HTTPRoute is in `default`, a `ReferenceGrant` is
required to authorize the cross-namespace backend reference.

Save the following to `catch-all.yaml`:

```yaml
# File: catch-all.yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata: {name: allow-default-to-error-pages, namespace: error-pages}
spec:
  from: [{group: gateway.networking.k8s.io, kind: HTTPRoute, namespace: default}]
  to: [{group: "", kind: Service, name: error-pages}]
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata: {name: catch-all, namespace: default}
spec:
  parentRefs: [{name: nginx, namespace: default}]
  rules:
    - matches: [{path: {type: PathPrefix, value: /}}]
      backendRefs: [{name: error-pages, namespace: error-pages, port: 8080}]
```

```shell
$ kubectl apply -f catch-all.yaml
referencegrant.gateway.networking.k8s.io/allow-default-to-error-pages created
httproute.gateway.networking.k8s.io/catch-all created
```

Requests to unknown hostnames now return a styled 404:

```shell
$ curl -s http://unknown.localtest.me/
Error 404: Not Found
The server can not find the requested page
```

```shell
$ curl -s -H "Accept: application/json" http://unknown.localtest.me/
{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}
```

Unmatched hostnames always return 404 - there is no backend to produce any other status code.

### Cleanup

Delete the kind cluster to remove all resources at once:

```shell
kind delete cluster --name error-pages-test-cluster
```
