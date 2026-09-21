#!/usr/bin/env bash

# Verifies Slack webhook URLs by sending a request with a deliberately wrong parameter type.
# A valid webhook returns HTTP 400 with {"ok":false,"error":"invalid_workflow_input"},
# which confirms the URL exists and is reachable without triggering an actual notification.
#
# SLACK_RELEASE_WEBHOOK expects "release" as a string — passing an object forces the 400.
# SLACK_WEBHOOK_URL expects "repository" as a string — same approach.

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
  BODY=$(echo "${RESPONSE}" | head -n-1)
  echo "HTTP status: ${HTTP_STATUS}"
  echo "Body: ${BODY}"

  ERROR=$(echo "${BODY}" | jq -r '.error // empty' 2>/dev/null)
  if [ "${HTTP_STATUS}" -eq 400 ] && [ "${ERROR}" = "invalid_workflow_input" ]; then
    echo "PASS: ${name} is valid"
  else
    >&2 echo "FAIL: ${name} returned unexpected response"
    exit 1
  fi
  end_group
}

check_webhook "SLACK_RELEASE_WEBHOOK" "${SLACK_RELEASE_WEBHOOK}" '{"release": {"this_should_be_a_string_not_an_object": true}}'
check_webhook "SLACK_WEBHOOK_URL" "${SLACK_WEBHOOK_URL}" '{"repository": {"this_should_be_a_string_not_an_object": true}}'
