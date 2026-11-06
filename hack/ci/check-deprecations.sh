#!/usr/bin/env bash

# Fetches deprecated API usage from the cluster API server metrics, then checks
# that only known/allowlisted resources appear. Exits with 1 if any unexpected
# deprecated API usage is found.
#
# Required tools: kubectl, prom2json, jq
# KUBECONFIG must be set (or the default context used) before calling this script.

set -eou pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ALLOWLIST_FILE="${script_dir}/allowlisted-deprecated-apis.txt"

IGNORED=$(grep -v '^#' "${ALLOWLIST_FILE}" | grep -v '^$' | jq -Rn '[inputs]')

echo "Allowlisted deprecated APIs:"
echo "${IGNORED}"

echo "Fetching deprecated API usage from cluster metrics..."
kubectl get --raw /metrics | prom2json | \
  jq '.[] | select(.name=="apiserver_requested_deprecated_apis").metrics[].labels' \
  > deprecations.txt

echo "Deprecated APIs reported by the cluster:"
cat deprecations.txt

echo "Deprecated APIs not in the allowlist (must be empty):"
NON_ALLOWLISTED=$(jq "select(.resource as \$r | ${IGNORED} | index(\$r) | not)" < deprecations.txt)
echo "${NON_ALLOWLISTED}"

if [[ -n "${NON_ALLOWLISTED}" ]]; then
  echo "ERROR: unexpected deprecated API usage detected"
  exit 1
fi

echo "OK: no unexpected deprecated API usage detected"
