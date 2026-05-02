# Use error pages with ingress-nginx in Kubernetes

> [!IMPORTANT]
> I am not a Kubernetes expert. This guide is **not a production-ready solution** - it is a working starting point
> and a cheat sheet for `error-pages` users. Everything described here has been tested and works, but every real
> deployment is different: validate the configuration, review the security implications, and adapt it to your own
> environment before using it in production.
>
> Contributions and improvements are very welcome - feel free to open a PR and I will happily accept it.

> [!WARNING]
> `kubernetes/ingress-nginx` was archived on March 24, 2026 and is no longer maintained. The final release is
> `controller-v1.15.1`. For new projects, consider migrating to the [Gateway API](https://gateway-api.sigs.k8s.io/)
> with a supported controller such as NGINX Gateway Fabric, Envoy Gateway, or Cilium. This guide targets the final
> stable release and still works as-is.

This guide wires `error-pages` as the **global default backend** of ingress-nginx. When any backend returns a
status code listed in `custom-http-errors`, the controller intercepts the response and forwards it to `error-pages`,
which returns a styled page in the format the client requested.

## Local cluster setup

Follow the [kind guide](install_k8s_kind.md) to install the prerequisites.

## Install ingress-nginx

The Helm repository remains available even though the project is archived:

```shell
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx && helm repo update
...Successfully got an update from the "ingress-nginx" chart repository
Update Complete. ⎈Happy Helming!⎈
```

Install the controller with kind-compatible settings:

```shell
$ helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --version 4.15.1 \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.hostPort.enabled=true \
  --set controller.service.type=LoadBalancer \
  --set controller.updateStrategy.type=RollingUpdate \
  --set controller.updateStrategy.rollingUpdate.maxUnavailable=1 \
  --set 'controller.nodeSelector.kubernetes\.io/os=linux' \
  --set 'controller.tolerations[0].key=node-role.kubernetes.io/master' \
  --set 'controller.tolerations[0].effect=NoSchedule' \
  --set 'controller.tolerations[0].operator=Exists' \
  --set 'controller.tolerations[1].key=node-role.kubernetes.io/control-plane' \
  --set 'controller.tolerations[1].effect=NoSchedule' \
  --set 'controller.tolerations[1].operator=Exists' \
  --wait \
  --timeout=90s

Release "ingress-nginx" does not exist. Installing it now.
NAME: ingress-nginx
LAST DEPLOYED: ...
NAMESPACE: ingress-nginx
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
TEST SUITE: None
```

Two kind-specific settings are worth explaining:

* `controller.updateStrategy.rollingUpdate.maxUnavailable=1` - each ingress pod binds port 80 directly on the host
  via `hostPort`. Two pods cannot hold the same host port at the same time. Setting `maxUnavailable: 1` allows the controller
  to terminate the old pod before starting a new one, avoiding port conflicts (this matches the upstream kind values).
* `controller.service.type=LoadBalancer` - matches the official kind configuration. Without `cloud-provider-kind`,
  the Service external IP stays `<pending>`, but that's fine here: traffic reaches the ingress controller via
  `hostPort` → `extraPortMappings`, not through the Service.

Verify the controller pod is ready:

```shell
$ kubectl get pods --namespace ingress-nginx
NAME                                        READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-5d4cf94bdc-xxxxx   1/1     Running   0          60s
```

Verify ingress-nginx is reachable on port `80`. There are no Ingress rules yet, so the controller returns its own
built-in 404 - that's expected at this stage:

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
  ports: [{port: 80, targetPort: 8080}]
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata: {name: httpbin}
spec:
  ingressClassName: nginx
  rules:
  - host: httpbin.localtest.me
    http: {paths: [{path: /, pathType: Prefix, backend: {service: {name: httpbin, port: {number: 80}}}}]}
```

`localtest.me` is a public DNS wildcard: `*.localtest.me` always resolves to `127.0.0.1`, so this works offline and
doesn't require an `/etc/hosts` entry.

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
returns the requested status code with an empty body - no message, no description, nothing useful for an end user:

```shell
$ curl -si http://httpbin.localtest.me/status/404
HTTP/1.1 404 Not Found
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
Date: ...
Content-Type: text/plain; charset=utf-8
Content-Length: 0
Connection: keep-alive
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
```

Requests that don't match any Ingress rule are served by ingress-nginx's built-in fallback:

```shell
$ curl -s http://unknown.localtest.me/ | head -n 5
<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx</center>
```

## 🔥 Install error-pages

Setting **`config.sendSameHttpCode=true`** is critical for the ingress-nginx integration - error-pages must return
the same HTTP status code as the error it renders, not `200`. Otherwise, ingress-nginx will proxy the `200`
response to the client and the actual error code will be lost.

Optionally, you can set **`config.showDetails=true`**. This tells error-pages to read Kubernetes-specific headers
that ingress-nginx includes with each intercepted request: `X-Original-URI`, `X-Namespace`, `X-Ingress-Name`,
`X-Service-Name`, `X-Service-Port`. Templates that support it will render this context alongside the error.

> [!NOTE]
> Get the latest chart version from [ArtifactHub](https://artifacthub.io/packages/helm/error-pages/error-pages)
> or the [GitHub releases](https://github.com/tarampampam/error-pages/releases) page.

```shell
$ helm install error-pages oci://ghcr.io/tarampampam/error-pages/charts/error-pages \
  --version X.Y.Z \
  --namespace error-pages \
  --create-namespace \
  --set config.sendSameHttpCode=true \
  --set config.showDetails=true \
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

## Wire ingress-nginx to error-pages

Two things need to be configured on the ingress-nginx controller:

* **`custom-http-errors`** (ConfigMap) - a list of status codes that ingress-nginx intercepts and forwards to the
  default backend instead of passing them through to the client.
* **`default-backend-service`** (controller argument) - the `namespace/service` that receives intercepted requests.
  Changing this argument triggers a pod restart.

Apply both changes with a single `helm upgrade --reuse-values`, which merges new values on top of the existing ones
and preserves all kind-specific settings:

```shell
$ helm upgrade ingress-nginx ingress-nginx/ingress-nginx \
  --version 4.15.1 \
  --namespace ingress-nginx \
  --reuse-values \
  --wait \
  --timeout=90s \
  --values - <<'EOF'
controller:
  config:
    custom-http-errors: "400,401,403,404,405,408,409,410,429,500,502,503,504"
  extraArgs:
    default-backend-service: "error-pages/error-pages"
EOF

Release "ingress-nginx" has been upgraded. Happy Helming!
NAME: ingress-nginx
LAST DEPLOYED: ...
NAMESPACE: ingress-nginx
STATUS: deployed
REVISION: 2
DESCRIPTION: Upgrade complete
```

## Verify custom error pages

The same requests that previously returned empty responses now return styled error pages. `ingress-nginx` intercepts
the upstream error, passes `X-Code` and the client's `Accept` header to `error-pages` as `X-Format`, and proxies the
styled response back to the client:

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
  "description": "The server can not find the requested page",
  "details": {
    "host": "",
    "original_uri": "/status/404",
    "forwarded_for": "172.20.0.1",
    "namespace": "default",
    "ingress_name": "httpbin",
    "service_name": "httpbin",
    "service_port": "80",
    "request_id": "8da6beeca4da9e373b2b0290e49bfac6",
    "timestamp": 1777573521
  }
}
```

```shell
$ curl -s -H "Accept: application/xml" http://httpbin.localtest.me/status/503
<?xml version="1.0" encoding="utf-8"?>
<error>
  <code>503</code>
  <message>Service Unavailable</message>
  <description>The server is temporarily overloading or down</description>
  <details>
    <host></host>
    <originalURI>/status/503</originalURI>
    <forwardedFor>172.20.0.1</forwardedFor>
    <namespace>default</namespace>
    <ingressName>httpbin</ingressName>
    <serviceName>httpbin</serviceName>
    <servicePort>80</servicePort>
    <requestID>829047aa49889b0a66ecc114febf39f1</requestID>
    <timestamp>1777573541</timestamp>
  </details>
</error>
```

Unmatched routes now go through `error-pages` as well, instead of the built-in fallback:

```shell
$ curl -s -H "Accept: application/json" http://unknown.localtest.me/
{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page",
  "details": {
    "host": "",
    "original_uri": "/",
    "forwarded_for": "172.20.0.1",
    "namespace": "",
    "ingress_name": "",
    "service_name": "",
    "service_port": "",
    "request_id": "f30b8b3ff3a988704753592fd89c2e32",
    "timestamp": 1777573580
  }
}
```

### Cleanup

Delete the kind cluster to remove all resources at once:

```shell
kind delete cluster --name error-pages-test-cluster
```
