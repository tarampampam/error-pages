# image page: <https://hub.docker.com/_/node>
FROM node:16.9-alpine as builder

# use directory with application sources by default
WORKDIR /app

# copy files, that required for the dependencies installing
COPY ./yarn.lock ./package.json ./

# install nodejs dependencies
RUN yarn install --frozen-lockfile --no-progress --non-interactive

# copy all application sources
COPY . .

# build the frontend application
RUN yarn build

# generate error pages
RUN ./dist/index.js build -c ./error-pages.yml ./error_pages

# create a directory for the future root filesystem
WORKDIR /tmp/rootfs

# prepare the root filesystem
RUN set -x \
    && mkdir -p ./bin ./etc ./home/runtime ./opt \
    && echo 'runtime:x:10001:10001::/home/runtime:/sbin/nologin' > ./etc/passwd \
    && echo 'runtime:x:10001:' > ./etc/group \
    && mv /app/caddy.json ./etc/caddy.json \
    && mv /app/error-pages.yml ./etc/error-pages.yml \
    && mv /app/dist ./app \
    && mv /app/error_pages ./opt/html \
    && chown -R 10001:10001 ./home/runtime

# use distroless node for the result image: <https://github.com/astefanutti/scratch-node>
FROM ghcr.io/astefanutti/scratch-node:16.9 as runtime

LABEL \
    # Docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
    org.opencontainers.image.title="error-pages" \
    org.opencontainers.image.description="Static server error pages in docker image" \
    org.opencontainers.image.url="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.source="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.vendor="tarampampam" \
    org.opencontainers.image.licenses="MIT"

# install curl
COPY --from=tarampampam/curl:7.78.0 /bin/curl /bin/curl

# install caddy file server (image page: <https://hub.docker.com/_/caddy>)
COPY --from=caddy:2.4.5-alpine /usr/bin/caddy /bin/caddy

# import the root filesystem
COPY --from=builder /tmp/rootfs /

ENV \
  TEMPLATE_NAME=ghost \
  CONFIG_FILE=/etc/error-pages.yml \
  TEMPLATES_DIR=/opt/html

# use an unprivileged user
USER runtime:runtime

# Docs: <https://docs.docker.com/engine/reference/builder/#healthcheck>
HEALTHCHECK --interval=15s --timeout=2s --retries=2 --start-period=2s CMD [ \
    "curl", "--fail", "--user-agent", "internal/healthcheck", "http://127.0.0.1:8080/health/live" \
]

ENTRYPOINT ["/bin/node", "/app/index.js", "init", "--log-json", "--"]

CMD ["/bin/caddy", "run", "-config", "/etc/caddy.json"]
