#!/usr/bin/env bash

# Verifies Slack webhook URLs by sending a real notification payload with dummy values.
# A reachable, correctly configured webhook returns HTTP 200.

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars SLACK_RELEASE_WEBHOOK SLACK_WEBHOOK_URL

check_webhook() {
  local name="$1"
  local url="$2"
  local payload="$3"

  start_group "Check ${name}"
  RESPONSE=$(curl --silent --write-out "\n%{http_code}" \
    --request POST "${url}" \
    --header "Content-Type: application/json" \
    --data "${payload}")
  HTTP_STATUS=$(echo "${RESPONSE}" | tail -n1)
  echo "HTTP status: ${HTTP_STATUS}"

  if [ "${HTTP_STATUS}" -eq 200 ]; then
    echo "PASS: ${name} is valid"
  else
    >&2 echo "FAIL: ${name} returned unexpected response"
    exit 1
  fi
  end_group
}

check_webhook "SLACK_RELEASE_WEBHOOK" "${SLACK_RELEASE_WEBHOOK}" \
  '{"repository":"webhook-test","release":"0.0.0","status":"success","actor":"webhook-test","email":"webhook-test@test.com","link":"https://webhook-test.example.com"}'

check_webhook "SLACK_WEBHOOK_URL" "${SLACK_WEBHOOK_URL}" \
  '{"repository":"webhook-test","workflow":"webhook-test","email":"webhook-test@test.com","attempt":"1","link":"https://webhook-test.example.com"}'
