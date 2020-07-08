# Image page: <https://hub.docker.com/_/node>
FROM node:12.16.2-alpine as builder

WORKDIR /src

COPY . .

RUN set -x \
    && yarn install --frozen-lockfile \
    && ./bin/generator.js -c ./config.json -o ./out

# Image page: <https://hub.docker.com/_/nginx>
FROM nginx:1.18-alpine

COPY --from=builder --chown=nginx /src/docker/docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=builder --chown=nginx /src/docker/nginx-server.conf /etc/nginx/conf.d/default.conf
COPY --from=builder --chown=nginx /src/static /opt/html
COPY --from=builder --chown=nginx /src/out /opt/html

ENTRYPOINT ["/docker-entrypoint.sh"]

CMD ["nginx", "-g", "daemon off;"]
