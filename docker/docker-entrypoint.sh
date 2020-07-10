#!/usr/bin/env sh
set -e

TEMPLATE_NAME=${TEMPLATE_NAME:-ghost} # string|empty

if [ -n "$TEMPLATE_NAME" ]; then
  echo "$0: set pages for template '$TEMPLATE_NAME' as default (make accessible in root directory)";

  if [ ! -d "/opt/html/$TEMPLATE_NAME" ]; then
    (>&2 echo "$0: template '$TEMPLATE_NAME' was not found!"); exit 1;
  fi;

  ln -f -s "/opt/html/$TEMPLATE_NAME/"* /opt/html;

  # next directory is required for easy nginx `error_page` usage
  mkdir /opt/html/nginx-error-pages;
  ln -f -s "/opt/html/$TEMPLATE_NAME/"* /opt/html/nginx-error-pages;
fi;

exec "$@"
