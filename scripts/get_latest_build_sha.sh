#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

last_build_id=$(curl -f -L -X GET https://storage.googleapis.com/kyma-prow-logs/logs/post-api-gateway-manager-build/latest-build.txt)
if [ -z "${last_build_id}" ]; then
  >&2 echo "Unable to retrieve last build ID"
  exit 1
fi

if ! command -v jq &> /dev/null; then
    sudo apt-get install jq
fi

finished=$(curl -f -L -X GET https://storage.googleapis.com/kyma-prow-logs/logs/post-api-gateway-manager-build/${last_build_id}/finished.json)
if [ -z "${finished}" ]; then
  >&2 echo "Unable to retrieve finished logs"
  exit 1
fi

result=$(echo ${finished} | jq -r '.result')
if [ "${result}" != "SUCCESS" ]; then
  >&2 echo "Last build was not successful"
  exit 1
fi

started=$(curl -f -L -X GET https://storage.googleapis.com/kyma-prow-logs/logs/post-api-gateway-manager-build/${last_build_id}/started.json)
if [ -z "${started}" ]; then
  >&2 echo "Unable to retrieve started logs"
  exit 1
fi

sha=$(echo ${started} | jq -r '.["repo-commit"]')
if [ -z "${sha}" ]; then
  >&2 echo "Unable to extract repo-commit from logs"
  exit 1
fi

echo $sha
