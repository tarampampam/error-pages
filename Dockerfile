# syntax=docker/dockerfile:1

# -✂- this stage is used to compile the application -------------------------------------------------------------------
FROM docker.io/library/golang:1.24-alpine AS compile

# can be passed with any prefix (like `v1.2.3@GITHASH`), e.g.: `docker build --build-arg "APP_VERSION=v1.2.3" .`
ARG APP_VERSION="undefined@docker"

# copy the source code
COPY . /src

WORKDIR /src

RUN set -x \
    && go generate -skip readme ./... \
    && CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X gh.tarampamp.am/error-pages/internal/appmeta.version=${APP_VERSION}" \
      -o /tmp/error-pages \
      ./cmd/error-pages/ \
    && /tmp/error-pages --version \
    && /tmp/error-pages -h

WORKDIR /tmp/rootfs

# prepare rootfs for runtime
RUN set -x \
    && mkdir -p ./etc/ssl/certs ./bin \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > ./etc/passwd \
    && echo 'appuser:x:10001:' > ./etc/group \
    && cp /etc/ssl/certs/ca-certificates.crt ./etc/ssl/certs/ \
    && mv /tmp/error-pages ./bin/error-pages \
    && chmod 755 ./bin/error-pages

WORKDIR /tmp/rootfs/opt

# generate static error pages (for use inside other Docker images, for example)
RUN set -x \
    && mkdir ./html \
    && ./../bin/error-pages build --index --target-dir ./html \
    && ls -l ./html

# -✂- and this is the final stage (an empty filesystem is used) -------------------------------------------------------
FROM scratch AS runtime

ARG APP_VERSION="undefined@docker"

LABEL \
    # docs: https://github.com/opencontainers/image-spec/blob/master/annotations.md
    org.opencontainers.image.title="error-pages" \
    org.opencontainers.image.description="Pretty server's error pages" \
    org.opencontainers.image.url="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.source="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.vendor="tarampampam" \
    org.opencontainers.version="$APP_VERSION" \
    org.opencontainers.image.licenses="MIT"

# import from builder
COPY --from=compile /tmp/rootfs /

# use an unprivileged user
USER 10001:10001

WORKDIR /opt

# to find out which environment variables and CLI arguments are supported by the application, run the app
# with the `--help` flag or refer to the documentation at https://github.com/tarampampam/error-pages#readme

ENV LOG_LEVEL="warn" \
    LOG_FORMAT="json"

# docs: https://docs.docker.com/reference/dockerfile/#healthcheck
HEALTHCHECK --interval=10s --start-interval=1s --start-period=2s --timeout=1s CMD ["/bin/error-pages", "healthcheck"]

ENTRYPOINT ["/bin/error-pages"]

CMD ["serve"]
