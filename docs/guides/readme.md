# 🛠 Integration guides

- **Standalone** server or builder usage:
  - [Start the HTTP server with a custom template (theme)](theming.md)
  - [Generate error pages using built-in or custom templates](builder.md)

- With **Nginx**:
  - [Customize error pages in your own Nginx Docker image](nginx_image.md)
  - [Use Nginx as a reverse proxy with custom error pages](nginx_upstream.md)

- With **Caddy**:
  - [Customize error pages in your own Caddy Docker image](caddy_image.md)
  - [Use Caddy as a reverse proxy with custom error pages](caddy_upstream.md)

- With **Traefik**:
  - [Use Traefik with a local Docker Compose setup](traefik_docker.md)

- With **Kubernetes**:
  - [Use error-pages as the **ingress-nginx** default backend](k8s_ingress_nginx.md)
  - [Use error-pages with **Traefik** in Kubernetes](k8s_ingress_traefik.md)
  - [Use error-pages with **NGINX Gateway Fabric** in Kubernetes](k8s_ingress_ngf.md)
  - [Use error-pages with **Envoy Gateway** in Kubernetes](k8s_ingress_envoy.md)
  - [Use error pages with **HAProxy Ingress** in Kubernetes](k8s_ingress_haproxy.md)
