# syntax=docker/dockerfile:1

# -✂- this stage is used to develop and build the application locally -------------------------------------------------
FROM docker.io/library/golang:1.22-bookworm AS develop

# use the /var/tmp as the GOPATH to reuse the modules cache
ENV GOPATH="/var/tmp"

RUN set -x \
    # renovate: source=github-releases name=abice/go-enum
    && ABICE_GOENUM_VERSION="0.6.0" \
    && GOBIN=/bin go install "github.com/abice/go-enum@v${ABICE_GOENUM_VERSION}" \
    && GOBIN=/bin go install golang.org/x/tools/cmd/goimports@latest \
    && GOBIN=/bin go install gotest.tools/gotestsum@latest \
    && go clean -cache -modcache \
    # renovate: source=github-releases name=golangci/golangci-lint
    && GOLANGCI_LINT_VERSION="1.59.1" \
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
    go mod download -x

# -✂- this stage is used to compile the application -------------------------------------------------------------------
FROM develop AS compile

# can be passed with any prefix (like `v1.2.3@GITHASH`), e.g.: `docker build --build-arg "APP_VERSION=v1.2.3" .`
ARG APP_VERSION="undefined@docker"

RUN --mount=type=bind,source=.,target=/src set -x \
    && go generate ./... \
    && CGO_ENABLED=0 LDFLAGS="-s -w -X gh.tarampamp.am/error-pages/internal/version.version=${APP_VERSION}" \
      go build -trimpath -ldflags "${LDFLAGS}" -o /tmp/error-pages ./cmd/error-pages/ \
    && /tmp/error-pages --version \
    && /tmp/error-pages -h

# -✂- this stage is used to prepare the runtime fs --------------------------------------------------------------------
FROM docker.io/library/alpine:3.20 AS rootfs

WORKDIR /tmp/rootfs

# prepare rootfs for runtime
RUN --mount=type=bind,source=.,target=/src set -x \
    && mkdir -p ./etc ./bin ./opt/html \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > ./etc/passwd \
    && echo 'appuser:x:10001:' > ./etc/group \
    && cp -rv /src/templates ./opt/templates \
    && rm -v ./opt/templates/*.md \
    && cp -rv /src/error-pages.yml ./opt/error-pages.yml

# take the binary from the compile stage
COPY --from=compile /tmp/error-pages ./bin/error-pages

WORKDIR /tmp/rootfs/opt

# generate static error pages (for usage inside another docker images, for example)
RUN set -x \
    && ./../bin/error-pages --verbose build --config-file ./error-pages.yml --index ./html \
    && ls -l ./html

# -✂- and this is the final stage (an empty filesystem is used) -------------------------------------------------------
FROM scratch AS runtime

ARG APP_VERSION="undefined@docker"

LABEL \
    # docs: https://github.com/opencontainers/image-spec/blob/master/annotations.md
    org.opencontainers.image.title="error-pages" \
    org.opencontainers.image.description="Static server error pages in the docker image" \
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

ENV LISTEN_PORT="8080" \
    TEMPLATE_NAME="ghost" \
    DEFAULT_ERROR_PAGE="404" \
    DEFAULT_HTTP_CODE="404" \
    SHOW_DETAILS="false" \
    DISABLE_L10N="false" \
    READ_BUFFER_SIZE="2048"

# docs: https://docs.docker.com/reference/dockerfile/#healthcheck
HEALTHCHECK --interval=10s --start-interval=1s --start-period=5s --timeout=2s CMD [\
  "/bin/error-pages", "--log-json", "healthcheck" \
]

ENTRYPOINT ["/bin/error-pages"]

CMD ["--log-json", "serve"]
