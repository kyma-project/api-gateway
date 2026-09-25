#!/usr/bin/env bash

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars GARDENER_TOKEN GARDENER_PROJECT_NAME

start_group "Verify GARDENER_TOKEN against Gardener API"
HTTP_STATUS=$(curl --silent --output /dev/null --write-out "%{http_code}" \
  --header "Authorization: Bearer ${GARDENER_TOKEN}" \
  "https://api.live.gardener.cloud.sap/apis/core.gardener.cloud/v1beta1/namespaces/garden-${GARDENER_PROJECT_NAME}/shoots")
echo "HTTP status: ${HTTP_STATUS}"
if [ "${HTTP_STATUS}" -ne 200 ]; then
  >&2 echo "FAIL: Gardener API returned HTTP ${HTTP_STATUS} — token may be expired or invalid"
  exit 1
fi
end_group

echo "PASS: GARDENER_TOKEN is valid"
