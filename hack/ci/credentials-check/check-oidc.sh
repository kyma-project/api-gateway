#!/usr/bin/env bash

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars OIDC_ISSUER_URL CLIENT_ID CLIENT_SECRET

start_group "Fetch token_endpoint from OIDC discovery document"
WELL_KNOWN_URL="${OIDC_ISSUER_URL}/.well-known/openid-configuration"
TOKEN_ENDPOINT=$(curl --silent --show-error --fail "${WELL_KNOWN_URL}" \
  | grep -o '"token_endpoint":"[^"]*"' | cut -d'"' -f4)
if [ -z "${TOKEN_ENDPOINT}" ]; then
  >&2 echo "FAIL: could not retrieve token_endpoint from ${WELL_KNOWN_URL}"
  exit 1
fi
echo "token_endpoint: ${TOKEN_ENDPOINT}"
end_group

start_group "Request token using client_credentials grant"
HTTP_STATUS=$(curl --silent --output /dev/null --write-out "%{http_code}" \
  --request POST "${TOKEN_ENDPOINT}" \
  --data-urlencode "grant_type=client_credentials" \
  --data-urlencode "client_id=${CLIENT_ID}" \
  --data-urlencode "client_secret=${CLIENT_SECRET}")
echo "HTTP status: ${HTTP_STATUS}"
if [ "${HTTP_STATUS}" -ne 200 ]; then
  >&2 echo "FAIL: token request returned HTTP ${HTTP_STATUS}"
  exit 1
fi
end_group

echo "PASS: OIDC credentials are valid"
