# syntax=docker/dockerfile:1

# -✂- this stage is used to compile the application -------------------------------------------------------------------
FROM docker.io/library/golang:1.26.5-alpine AS compile

# can be passed with any prefix (like `v1.2.3@GITHASH`), e.g.: `(podman|docker) build --build-arg "APP_VERSION=v1.2.3" .`
ARG APP_VERSION="undefined@docker"

# copy the source code
COPY . /src

WORKDIR /src

RUN set -x \
    && LDFLAGS="-s -w -X gh.tarampamp.am/error-pages/v4/internal/appmeta.version=${APP_VERSION}" \
    && go generate -skip readme ./... \
    && CGO_ENABLED=0 go build -trimpath -ldflags "${LDFLAGS}" -o /tmp/error-pages ./cmd/error-pages/ \
    && CGO_ENABLED=0 go build -trimpath -ldflags "${LDFLAGS}" -o /tmp/builder ./cmd/builder/ \
    && /tmp/error-pages --version \
    && /tmp/error-pages -h \
    && /tmp/builder --version \
    && /tmp/builder -h

# prepare the common rootfs
WORKDIR /tmp/rootfs-base
RUN set -x \
    && mkdir -p ./etc/ssl/certs ./bin ./tmp \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > ./etc/passwd \
    && echo 'appuser:x:10001:' > ./etc/group \
    && cp /etc/ssl/certs/ca-certificates.crt ./etc/ssl/certs/ \
    && chmod 1777 ./tmp

# prepare separate rootfs for the server
WORKDIR /tmp/rootfs/server
RUN set -x \
    && cp -a /tmp/rootfs-base/. . \
    && mv /tmp/error-pages ./bin/error-pages \
    && chmod 555 ./bin/error-pages

# add super-lightweight HTTP checking tool to use it in the healthcheck
# docs: https://github.com/tarampampam/microcheck
COPY --from=ghcr.io/tarampampam/microcheck:1 /bin/httpcheck /tmp/rootfs/server/bin/httpcheck

# and prepare separate rootfs for the builder (plus generate static error pages)
WORKDIR /tmp/rootfs/builder
RUN set -x \
    && cp -a /tmp/rootfs-base/. . \
    && mv /tmp/builder ./bin/builder \
    && chmod 555 ./bin/builder \
    && mkdir -p ./opt/html \
    && chmod 755 ./opt/html \
    && ./bin/builder --target-dir ./opt/html \
    && ls -l ./opt/html

# -✂- this stage is used to prepare and reuse metadata for both final stages (server and builder) ---------------------
FROM scratch AS runtime-base

ARG APP_VERSION="undefined@docker"

# docs: https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL \
    org.opencontainers.image.title="error-pages" \
    org.opencontainers.image.description="Tiny HTTP server that serves clean, themeable, localized HTTP error pages" \
    org.opencontainers.image.url="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.source="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.vendor="tarampampam" \
    org.opencontainers.version="$APP_VERSION" \
    org.opencontainers.image.licenses="MIT"

# use an unprivileged user by dedault
USER 10001:10001

# -✂- this is the final stage for the builder -------------------------------------------------------------------------
# to build the builder image, use: `(podman|docker) build --target=builder ...`
FROM runtime-base AS builder

# import rootfs for the builder from the compile stage
COPY --from=compile /tmp/rootfs/builder /

ENTRYPOINT ["/bin/builder"]

# -✂- and this is the final stage for the server ----------------------------------------------------------------------
# to build the server image, use: `(podman|docker) build --target=server ...`
FROM runtime-base AS server

# import rootfs for the server from the compile stage
COPY --from=compile /tmp/rootfs/server /

# to find out which environment variables and CLI arguments are supported by the application, run the app
# with the `--help` flag or refer to the documentation at https://github.com/tarampampam/error-pages#readme

ENV LOG_LEVEL="info" \
    LOG_FORMAT="json"

# docs: https://docs.docker.com/reference/dockerfile/#healthcheck
HEALTHCHECK --interval=10s --start-interval=1s --start-period=1s CMD [\
  "/bin/httpcheck", "--port-env", "HTTP_PORT", "http://127.0.0.1:8080/healthz"\
]

ENTRYPOINT ["/bin/error-pages"]
