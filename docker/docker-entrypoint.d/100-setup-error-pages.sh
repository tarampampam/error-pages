#!/usr/bin/env sh
set -e

# allows to use random template
if [ ! -z "$TEMPLATE_NAME" ] && ([ "$TEMPLATE_NAME" = "random" ] || [ "$TEMPLATE_NAME" = "RANDOM" ]); then
  # find all templates in directory (only template directories must be located in /opt/html)
  allowed_templates=$(find /opt/html/* -maxdepth 1 -type d ! -iname nginx-error-pages -exec basename {} \;);

  # pick random template name
  random_template_name=$(shuf -e -n1 $allowed_templates);

  echo "$0: Use '$random_template_name' as randomly selected template";

  TEMPLATE_NAME="$random_template_name"
fi;

TEMPLATE_NAME=${TEMPLATE_NAME:-ghost} # string|empty

echo "$0: Set pages for template '$TEMPLATE_NAME' as default (make accessible in root directory)";

# check for template existing
if [ ! -d "/opt/html/$TEMPLATE_NAME" ]; then
  echo >&3 "$0: Template '$TEMPLATE_NAME' was not found!";
  exit 1;
fi;

# allows "direct access" to the error pages using URLs like "/500.html"
ln -f -s "/opt/html/$TEMPLATE_NAME/"* /opt/html;

# on `docker restart` next directory keep existing: <https://github.com/tarampampam/error-pages/issues/3>
if [ -d /opt/html/nginx-error-pages ]; then
  rm -Rf /opt/html/nginx-error-pages;
fi;

# next directory is required for easy nginx `error_page` usage
mkdir /opt/html/nginx-error-pages;

# use error pages from the template as "native" nginx error pages
ln -f -s "/opt/html/$TEMPLATE_NAME/"* /opt/html/nginx-error-pages;

exit 0
