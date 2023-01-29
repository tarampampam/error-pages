# syntax=docker/dockerfile:1.2

# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.19-alpine as builder

# can be passed with any prefix (like `v1.2.3@GITHASH`), e.g.: `docker build --build-arg "APP_VERSION=v1.2.3@GITHASH" .`
ARG APP_VERSION="undefined@docker"

WORKDIR /src

COPY . .

# arguments to pass on each go tool link invocation
ENV LDFLAGS="-s -w -X github.com/tarampampam/error-pages/internal/version.version=$APP_VERSION"

RUN set -x \
    && go version \
    && CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o ./error-pages ./cmd/error-pages/ \
    && ./error-pages --version \
    && ./error-pages -h

WORKDIR /tmp/rootfs

# prepare rootfs for runtime
RUN set -x \
    && mkdir -p \
        ./etc \
        ./bin \
        ./opt/html \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > ./etc/passwd \
    && echo 'appuser:x:10001:' > ./etc/group \
    && mv /src/error-pages ./bin/error-pages \
    && mv /src/templates ./opt/templates \
    && rm ./opt/templates/*.md \
    && mv /src/error-pages.yml ./opt/error-pages.yml

WORKDIR /tmp/rootfs/opt

# generate static error pages (for usage inside another docker images, for example)
RUN set -x \
    && ./../bin/error-pages --verbose build --config-file ./error-pages.yml --index ./html \
    && ls -l ./html

# use empty filesystem
FROM scratch as runtime

ARG APP_VERSION="undefined@docker"

LABEL \
    # Docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
    org.opencontainers.image.title="error-pages" \
    org.opencontainers.image.description="Static server error pages in the docker image" \
    org.opencontainers.image.url="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.source="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.vendor="tarampampam" \
    org.opencontainers.version="$APP_VERSION" \
    org.opencontainers.image.licenses="MIT"

# Import from builder
COPY --from=builder /tmp/rootfs /

# Use an unprivileged user
USER 10001:10001

WORKDIR /opt

ENV LISTEN_PORT="8080" \
    TEMPLATE_NAME="ghost" \
    DEFAULT_ERROR_PAGE="404" \
    DEFAULT_HTTP_CODE="404" \
    SHOW_DETAILS="false" \
    DISABLE_L10N="false"

# Docs: <https://docs.docker.com/engine/reference/builder/#healthcheck>
HEALTHCHECK --interval=7s --timeout=2s CMD ["/bin/error-pages", "healthcheck", "--log-json"]

ENTRYPOINT ["/bin/error-pages"]

CMD ["serve", "--log-json"]
