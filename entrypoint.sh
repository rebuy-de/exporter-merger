#!/bin/sh

if [ -z "$COMMA_SEPARATED_URLS" ]; then
  echo Env var COMMA_SEPARATED_URLS not defined
  # Same behavior as before.
  exec /go/bin/exporter-merger $@
fi

(echo "exporters:"; echo "$COMMA_SEPARATED_URLS"|tr "," "\n"|xargs -n1 echo "- url:" ) > merger.yaml

echo merger.yaml:
cat merger.yaml
echo

exec /go/bin/exporter-merger --config-path merger.yaml $@
