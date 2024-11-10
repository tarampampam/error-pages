# syntax=docker/dockerfile:1

# -✂- this stage is used to develop and build the application locally -------------------------------------------------
FROM docker.io/library/golang:1.23-bookworm AS develop

# use the /var/tmp/go as the GOPATH to reuse the modules cache
ENV GOPATH="/var/tmp/go"

RUN set -x \
    # renovate: source=github-releases name=golangci/golangci-lint
    && GOLANGCI_LINT_VERSION="1.62.0" \
    && wget -O- -nv "https://cdn.jsdelivr.net/gh/golangci/golangci-lint@v${GOLANGCI_LINT_VERSION}/install.sh" \
      | sh -s -- -b /bin "v${GOLANGCI_LINT_VERSION}"

RUN set -x \
    # customize the shell prompt (for the bash)
    && echo "PS1='\[\033[1;36m\][go] \[\033[1;34m\]\w\[\033[0;35m\] \[\033[1;36m\]# \[\033[0m\]'" >> /etc/bash.bashrc

WORKDIR /src

# burn the modules cache
RUN \
    --mount=type=bind,source=go.mod,target=/src/go.mod \
    --mount=type=bind,source=go.sum,target=/src/go.sum \
    go mod download -x \
    && find "${GOPATH}" -type d -exec chmod 0777 {} \; \
    && find "${GOPATH}" -type f -exec chmod 0666 {} \;

# -✂- this stage is used to compile the application -------------------------------------------------------------------
FROM develop AS compile

# can be passed with any prefix (like `v1.2.3@GITHASH`), e.g.: `docker build --build-arg "APP_VERSION=v1.2.3" .`
ARG APP_VERSION="undefined@docker"

# copy the source code
COPY . /src

RUN set -x \
    && go generate ./... \
    && CGO_ENABLED=0 LDFLAGS="-s -w -X gh.tarampamp.am/error-pages/internal/appmeta.version=${APP_VERSION}" \
      go build -trimpath -ldflags "${LDFLAGS}" -o /tmp/error-pages ./cmd/error-pages/ \
    && /tmp/error-pages --version \
    && /tmp/error-pages -h

# -✂- this stage is used to prepare the runtime fs --------------------------------------------------------------------
FROM docker.io/library/alpine:3.20 AS rootfs

WORKDIR /tmp/rootfs

# prepare rootfs for runtime
RUN set -x \
    && mkdir -p ./etc/ssl/certs ./bin \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > ./etc/passwd \
    && echo 'appuser:x:10001:' > ./etc/group \
    && cp /etc/ssl/certs/ca-certificates.crt ./etc/ssl/certs/

# take the binary from the compile stage
COPY --from=compile /tmp/error-pages ./bin/error-pages

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
COPY --from=rootfs /tmp/rootfs /

# use an unprivileged user
USER 10001:10001

WORKDIR /opt

# to find out which environment variables and CLI arguments are supported by the application, run the app
# with the `--help` flag or refer to the documentation at https://github.com/tarampampam/error-pages#readme

ENV LOG_LEVEL="warn" \
    LOG_FORMAT="json"

# docs: https://docs.docker.com/reference/dockerfile/#healthcheck
HEALTHCHECK --interval=10s --start-interval=1s --start-period=5s --timeout=2s CMD ["/bin/error-pages", "healthcheck"]

ENTRYPOINT ["/bin/error-pages"]

CMD ["serve"]
