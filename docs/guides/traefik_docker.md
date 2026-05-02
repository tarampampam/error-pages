# Use Traefik with a local Docker Compose setup

Instead of writing a thousand words, let's just look at a single Compose file:

```yaml
# yaml-language-server: $schema=https://cdn.jsdelivr.net/gh/compose-spec/compose-spec@master/schema/compose-spec.json
# file: compose.yml

services:
  traefik:
    image: docker.io/library/traefik:v3.6
    command:
      # DEBUG logs every routing decision and middleware action - helpful during setup
      - --log.level=DEBUG
      # enable the Traefik dashboard - a web UI that shows routers, services, and middlewares
      - --api.dashboard=true
      # serve the dashboard and API without authentication on port 8080 (built-in);
      # NEVER use this in production - anyone who can reach port 8080 gets full control;
      # here it's safe because we only expose port 80 and rely on localtest.me routing
      - --api.insecure=true
      # tell Traefik to watch Docker for containers and read their labels as routing config
      - --providers.docker=true
      # do NOT automatically expose every container to the internet;
      # a container must have the label "traefik.enable=true" to be routed by Traefik
      - --providers.docker.exposedbydefault=false
      # create an entrypoint called "web" that listens on port 80 (plain HTTP);
      # the name "web" is arbitrary - it is referenced later in router labels
      - --entrypoints.web.address=:80

    ports: ['80:80/tcp']
    volumes: [/var/run/docker.sock:/var/run/docker.sock:ro]

    labels:
      # required: opt this container in (because exposedbydefault=false is set above)
      traefik.enable: true
      # Traefik dashboard: route requests to traefik.localtest.me -> the built-in dashboard UI;
      # localtest.me is a public DNS wildcard: `*.localtest.me` always resolves to 127.0.0.1,
      # so it works offline and needs no /etc/hosts entry
      traefik.http.routers.traefik.rule: Host(`traefik.localtest.me`)
      # "api@internal" is Traefik's built-in service that serves the dashboard and API
      traefik.http.routers.traefik.service: api@internal
      # attach this router to the "web" entrypoint defined above (port 80)
      traefik.http.routers.traefik.entrypoints: web
      # apply error-pages middleware to the dashboard too, so Traefik's own 4xx/5xx responses
      # also get the styled error page treatment
      traefik.http.routers.traefik.middlewares: error-pages-middleware

    depends_on:
      # do not start Traefik until error-pages is up and passing its health check
      error-pages: {condition: service_healthy}

  error-pages:
    image: ghcr.io/tarampampam/error-pages:4
    environment:
      # choose which built-in HTML template to use for error pages
      TEMPLATE_NAME: l7

    labels:
      # required: opt this container in
      traefik.enable: true

      # catch-all (fallback) router (matches ANY hostname):
      # this router acts as a fallback - when a request does not match any other registered
      # service, Traefik lands here and the errors middleware below rewrites the response into
      # a styled error page
      traefik.http.routers.error-pages-router.rule: HostRegexp(`.+`)
      # Traefik assigns automatic priority based on rule specificity (longer = higher priority);
      # `HostRegexp(.+)` would normally win over everything else, so we force a low priority (10)
      # to make sure specific Host(...) rules always win first
      traefik.http.routers.error-pages-router.priority: 10
      # attach to the "web" (HTTP) entrypoint on port 80
      traefik.http.routers.error-pages-router.entrypoints: web
      # apply the errors middleware to this catch-all router as well, so that error-pages own
      # responses (e.g., if it somehow returns 5xx) also go through the middleware
      traefik.http.routers.error-pages-router.middlewares: error-pages-middleware

      # errors middleware definitions:
      # intercept any upstream response with an HTTP status in the range 400–599
      traefik.http.middlewares.error-pages-middleware.errors.status: 400-599
      # forward intercepted responses to the "error-pages-service" defined below
      traefik.http.middlewares.error-pages-middleware.errors.service: error-pages-service
      # path template for the error page request; {status} is replaced with the actual HTTP
      # status code (e.g., 404, 503), so Traefik will request /404, /503, etc.
      traefik.http.middlewares.error-pages-middleware.errors.query: /{status}

      # service definition:
      # tell Traefik which port the error-pages container listens on internally; port 8080 is
      # the default HTTP port for this image (no need to publish it to the host)
      traefik.http.services.error-pages-service.loadbalancer.server.port: 8080

  nginx-or-any-another-service:
    # any regular service behind Traefik (replace this with your own app image)
    image: docker.io/library/nginx:1.29-alpine
    labels:
      # required: opt this container in
      traefik.enable: true
      # route requests for test.localtest.me to this nginx container
      traefik.http.routers.test-service.rule: Host(`test.localtest.me`)
      # attach to the "web" (HTTP) entrypoint
      traefik.http.routers.test-service.entrypoints: web
      # apply the errors middleware so that any 4xx/5xx response from nginx is replaced
      # with a styled error page from the error-pages service
      traefik.http.routers.test-service.middlewares: error-pages-middleware
```

After running `docker compose up` in the same directory as `compose.yml`, you can:

* Open the Traefik dashboard at [`traefik.localtest.me`](http://traefik.localtest.me/dashboard/#/)
* See a custom error page: [`http://traefik.localtest.me/foobar404`](http://traefik.localtest.me/foobar404)
* Open the nginx index page at [`test.localtest.me`](http://test.localtest.me/)
* See custom error pages for non-existent [pages](http://test.localtest.me/404) and even non-existent
  [domains](http://404.localtest.me/)

```shell
curl -s -H "Accept: application/json" http://traefik.localtest.me/foobar404
{
  "error": true,
  "code": 404,
  "message": "Not Found",
  "description": "The server can not find the requested page"
}

curl -s -H "X-Format: application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" http://test.localtest.me/404
<?xml version="1.0" encoding="utf-8"?>
<error>
  <code>404</code>
  <message>Not Found</message>
  <description>The server can not find the requested page</description>
</error>
```

Pretty neat, right? 🙂
