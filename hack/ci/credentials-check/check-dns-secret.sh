#!/usr/bin/env bash

# Verifies DNS_SECRET_JSON by activating the GCP service account it contains
# and listing Cloud DNS managed zones in the corresponding project.
# DNS_SA_BASE64 must be the base64-encoded GCP service account JSON
# (same encoding used by the e2e-test-gardener action).

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars DNS_SA_BASE64

start_group "Decode and validate service account JSON"
SA_JSON=$(echo "${DNS_SA_BASE64}" | base64 --decode)
if ! echo "${SA_JSON}" | jq empty 2>/dev/null; then
  >&2 echo "FAIL: DNS_SECRET_JSON does not decode to valid JSON"
  exit 1
fi

GCP_PROJECT=$(echo "${SA_JSON}" | jq -r '.project_id // empty')
if [ -z "${GCP_PROJECT}" ]; then
  >&2 echo "FAIL: service account JSON has no project_id field"
  exit 1
fi
end_group

start_group "Activate GCP service account"
SA_KEY_FILE=$(mktemp)
echo "${SA_JSON}" > "${SA_KEY_FILE}"
gcloud auth activate-service-account --key-file="${SA_KEY_FILE}" --quiet 2>/dev/null
rm -f "${SA_KEY_FILE}"
end_group

start_group "List Cloud DNS managed zones"
gcloud dns managed-zones list --project="${GCP_PROJECT}" --quiet > /dev/null
end_group

echo "PASS: DNS secret is valid and has access to Cloud DNS"
