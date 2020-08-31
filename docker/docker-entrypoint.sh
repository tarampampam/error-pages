#!/usr/bin/env sh
set -e

TEMPLATE_NAME=${TEMPLATE_NAME:-ghost} # string|empty

if [ -n "$TEMPLATE_NAME" ]; then
  echo "$0: set pages for template '$TEMPLATE_NAME' as default (make accessible in root directory)";

  if [ ! -d "/opt/html/$TEMPLATE_NAME" ]; then
    (>&2 echo "$0: template '$TEMPLATE_NAME' was not found!"); exit 1;
  fi;

  ln -f -s "/opt/html/$TEMPLATE_NAME/"* /opt/html;

  # on `docker restart` next directory keep existing: <https://github.com/tarampampam/error-pages/issues/3>
  if [ -d /opt/html/nginx-error-pages ]; then
    rm -Rf /opt/html/nginx-error-pages;
  fi;

  # next directory is required for easy nginx `error_page` usage
  mkdir /opt/html/nginx-error-pages;
  ln -f -s "/opt/html/$TEMPLATE_NAME/"* /opt/html/nginx-error-pages;
fi;

exec "$@"
