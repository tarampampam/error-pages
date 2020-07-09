#!/usr/bin/env sh
set -e

TEMPLATE_NAME=${TEMPLATE_NAME:-} # string|empty
DEFAULT_ERROR_CODE=${DEFAULT_ERROR_CODE:-404} # numeric

if [ -n "$TEMPLATE_NAME" ]; then
  echo "$0: set pages for template '$TEMPLATE_NAME' as default (make accessible in root directory)";

  if [ ! -d "/opt/html/$TEMPLATE_NAME" ]; then
    (>&2 echo "$0: template '$TEMPLATE_NAME' was not found!"); exit 1;
  fi;

  ln -f -s "/opt/html/$TEMPLATE_NAME/"* /opt/html;

  if [ -L "/opt/html/$DEFAULT_ERROR_CODE.html" ]; then
    echo "$0: set page with error code '$DEFAULT_ERROR_CODE' as default (index) page";

    cp -f "/opt/html/$DEFAULT_ERROR_CODE.html" /opt/html/index.html;
  else
    (>&2 echo "$0: cannot set page with error code '$DEFAULT_ERROR_CODE' as default (index) page!");
  fi;
fi;

exec "$@"
