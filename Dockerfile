# Image page: <https://hub.docker.com/_/node>
FROM node:15.14-alpine as builder

# copy required sources into builder image
COPY ./generator /src/generator
COPY ./config.json /src
COPY ./templates /src/templates
COPY ./docker /src/docker

# install generator dependencies
WORKDIR /src/generator
RUN yarn install --frozen-lockfile --no-progress --non-interactive

# run generator
WORKDIR /src
RUN ./generator/generator.js -c ./config.json -o ./out

# prepare rootfs for runtime
RUN mkdir /tmp/rootfs
WORKDIR /tmp/rootfs
RUN set -x \
    && mkdir -p \
        ./docker-entrypoint.d \
        ./etc/nginx/conf.d \
        ./opt \
    && mv /src/out ./opt/html \
    && echo -e "User-agent: *\nDisallow: /\n" > ./opt/html/robots.txt \
    && touch ./opt/html/favicon.ico \
    && mv /src/docker/docker-entrypoint.d/* ./docker-entrypoint.d \
    && mv /src/docker/nginx-server.conf ./etc/nginx/conf.d/default.conf

# Image page: <https://hub.docker.com/_/nginx>
FROM nginx:1.21-alpine as runtime

LABEL \
    # Docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
    org.opencontainers.image.title="error-pages" \
    org.opencontainers.image.description="Static server error pages in docker image" \
    org.opencontainers.image.url="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.source="https://github.com/tarampampam/error-pages" \
    org.opencontainers.image.vendor="tarampampam" \
    org.opencontainers.image.licenses="MIT"

# Import from builder
COPY --from=builder /tmp/rootfs /

RUN chown -R nginx:nginx /opt/html
