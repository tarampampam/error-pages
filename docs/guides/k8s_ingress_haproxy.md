# Use error pages with HAProxy Ingress in Kubernetes

> [!IMPORTANT]
> I am not a Kubernetes expert. This guide is **not a production-ready solution** - it is a working starting point
> and a cheat sheet for `error-pages` users. Everything described here has been tested and works, but every real
> deployment is different: validate the configuration, review the security implications, and adapt it to your own
> environment before using it in production.
>
> Contributions and improvements are very welcome - feel free to open a PR and I will happily accept it.

> [!WARNING]
> **HAProxy Ingress handles unmatched routes only**. Unlike ingress-nginx or Traefik, HAProxy Ingress does not have
> a built-in mechanism to intercept HTTP error responses from backends and replace them transparently. The
> `defaultBackendService` option wires `error-pages` as the fallback for **unmatched routes only** - requests to
> host/path combinations not covered by any Ingress rule. Errors returned by matched backends (for example, a
> deployment that responds with 500) pass through to the client unchanged.

This guide wires `error-pages` as the **default backend** of HAProxy Ingress. When a request arrives that no
Ingress rule matches, the controller forwards it to `error-pages`, which returns a styled page in the format the
client requested (HTML, JSON, XML, or plain text).

## Local cluster setup

Follow the [kind guide](install_k8s_kind.md) to install the prerequisites.

## Install HAProxy Ingress

```shell
$ helm repo add haproxy-ingress https://haproxy-ingress.github.io/charts && helm repo update
...Successfully got an update from the "haproxy-ingress" chart repository
Update Complete. ⎈Happy Helming!⎈
```

Install the controller as a `DaemonSet` so it binds ports `80` and `443` directly via `hostPort` on the kind node:

> [!NOTE]
> Get the latest chart version from the [GitHub releases](https://github.com/haproxy-ingress/charts/releases) page.

```shell
$ helm upgrade --install haproxy-ingress haproxy-ingress/haproxy-ingress \
  --version 0.16.0 \
  --namespace ingress-controller \
  --create-namespace \
  --set controller.kind=DaemonSet \
  --set controller.daemonset.useHostPort=true \
  --set controller.service.type=ClusterIP \
  --set controller.ingressClassResource.enabled=true \
  --set-string 'controller.nodeSelector.ingress-ready=true' \
  --set 'controller.tolerations[0].key=node-role.kubernetes.io/master' \
  --set 'controller.tolerations[0].effect=NoSchedule' \
  --set 'controller.tolerations[0].operator=Exists' \
  --set 'controller.tolerations[1].key=node-role.kubernetes.io/control-plane' \
  --set 'controller.tolerations[1].effect=NoSchedule' \
  --set 'controller.tolerations[1].operator=Exists' \
  --wait \
  --timeout=90s

Release "haproxy-ingress" does not exist. Installing it now.
NAME: haproxy-ingress
LAST DEPLOYED: ...
NAMESPACE: ingress-controller
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
```

A few kind-specific settings worth explaining:

* `controller.kind=DaemonSet` + `controller.daemonset.useHostPort=true` - the DaemonSet pod binds port 80 on the
  kind node directly via `hostPort`. Combined with the `extraPortMappings` in the kind cluster config, traffic
  flows: `localhost:80` → kind node port 80 → HAProxy Ingress pod port 80.
* `controller.service.type=ClusterIP` - the Service external IP is not needed here because traffic enters via
  `hostPort`, not through a LoadBalancer.
* `controller.ingressClassResource.enabled=true` - creates the `haproxy` IngressClass resource so that Ingress
  objects using `ingressClassName: haproxy` are picked up by this controller.
* `controller.nodeSelector.ingress-ready=true` - targets only nodes labelled `ingress-ready=true`, which is the
  label set in the kind cluster config for the control-plane node.

Verify the controller pod is ready:

```shell
$ kubectl get pods --namespace ingress-controller
NAME                     READY   STATUS    RESTARTS   AGE
haproxy-ingress-xxxxx    1/1     Running   0          30s
```

Verify HAProxy Ingress is reachable on port `80`. There are no Ingress rules yet, so the controller returns its
own built-in 404 - that is expected at this stage:

```shell
$ curl -si http://localhost/
HTTP/1.1 404 Not Found
content-type: text/html
cache-control: no-cache
content-length: 83

<html><body><h1>404 Not Found</h1>
The requested URL was not found.
</body></html>
```

## Deploy a test application

[go-httpbin](https://github.com/mccutchen/go-httpbin) is a small HTTP testing service. Its `/status/{code}`
endpoint returns any requested status code, which makes it handy for testing error page behaviour.

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
  ingressClassName: haproxy
  rules:
  - host: httpbin.localtest.me
    http: {paths: [{path: /, pathType: Prefix, backend: {service: {name: httpbin, port: {number: 80}}}}]}
```

`localtest.me` is a public DNS wildcard: `*.localtest.me` always resolves to `127.0.0.1`, so this works offline
and does not require an `/etc/hosts` entry.

```shell
$ kubectl apply -f httpbin.yaml && kubectl rollout status deployment/httpbin --timeout=60s
deployment.apps/httpbin created
service/httpbin created
ingress.networking.k8s.io/httpbin created
deployment "httpbin" successfully rolled out
```

Verify the application responds normally:

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
    "...": ["..."]
  },
  "method": "GET",
  "origin": "172.20.0.1",
  "url": "http://httpbin.localtest.me/get"
}
```

## Default error pages

Before `error-pages` is wired in, requests to unknown hosts are served by HAProxy Ingress's built-in fallback:

```shell
$ curl -si http://unknown.localtest.me/
HTTP/1.1 404 Not Found
content-type: text/html
content-length: 83
cache-control: no-cache

<html><body><h1>404 Not Found</h1>
The requested URL was not found.
</body></html>
```

Backend errors from matched routes pass through unmodified - go-httpbin's `/status/{code}` returns the requested
status code with an empty body:

```shell
$ curl -si http://httpbin.localtest.me/status/503
HTTP/1.1 503 Service Unavailable
access-control-allow-credentials: true
access-control-allow-origin: *
content-type: text/plain; charset=utf-8
date: ...
content-length: 0
```

## 🔥 Install error-pages

Setting **`config.sendSameHttpCode=true`** is required - `error-pages` must return the original error status code
(e.g. `404`) rather than `200`. Without it, HAProxy Ingress would forward the `200` response to the client and the
error code would be lost.

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

## Wire HAProxy Ingress to error-pages

Set `controller.defaultBackendService` to route all unmatched requests to `error-pages`. Changing this value
triggers a DaemonSet rollout:

```shell
$ helm upgrade haproxy-ingress haproxy-ingress/haproxy-ingress \
  --version 0.16.0 \
  --namespace ingress-controller \
  --reuse-values \
  --set controller.defaultBackendService=error-pages/error-pages \
  --wait \
  --timeout=90s

Release "haproxy-ingress" has been upgraded. Happy Helming!
NAME: haproxy-ingress
LAST DEPLOYED: ...
NAMESPACE: ingress-controller
STATUS: deployed
REVISION: 2
DESCRIPTION: Upgrade complete
```

## Verify custom error pages

Requests to unknown hosts now return styled pages from `error-pages`. HAProxy Ingress forwards the request -
including the original `Accept` header - to `error-pages`, which selects the appropriate format automatically:

```shell
$ curl -si -H "Accept: text/html" http://unknown.localtest.me/ | head -n 15
HTTP/1.1 404 Not Found
content-length: 52980
content-type: text/html; charset=utf-8
x-robots-tag: noindex, nofollow, nosnippet, noarchive
date: ...

<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="robots" content="nofollow,noarchive,noindex">
  <title>Not Found</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <!--  -->
  <meta name="title" content="404: Not Found">
```

```shell
$ curl -si -H "Accept: application/json" http://unknown.localtest.me/
HTTP/1.1 404 Not Found
content-length: 124
content-type: application/json; charset=utf-8
x-robots-tag: noindex, nofollow, nosnippet, noarchive
date: ...

{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}
```

Backend errors from matched routes remain unchanged:

```shell
$ curl -si http://httpbin.localtest.me/status/503
HTTP/1.1 503 Service Unavailable
access-control-allow-credentials: true
access-control-allow-origin: *
content-type: text/plain; charset=utf-8
date: ...
content-length: 0

(empty body - HAProxy Ingress does not intercept backend error responses)
```

## Cleanup

Delete the kind cluster to remove all resources at once:

```shell
kind delete cluster --name error-pages-test-cluster
```
