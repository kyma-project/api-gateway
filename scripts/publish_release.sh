#!/usr/bin/env bash

# This script publishes a release

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

release_id=$1

repository="${REPOSITORY:-kyma-project/api-gateway}"
github_api_repo_url="https://api.github.com/repos/${repository}"

echo "Publish release: repository: ${repository}: release ID: ${release_id}"

echo "Getting information about current release with ID = ${release_id}"
current_release=$(curl -s -S -f -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${github_api_repo_url}/releases/${release_id}" | \
  jq -r '.tag_name')
echo "Current release = ${current_release}"

echo "Getting latest release"
latest_release=$(curl -s -S -f -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${github_api_repo_url}/releases/latest" | \
  jq -r '.tag_name')
echo "Latest release = ${latest_release}"

expected_latest_release=$(printf '%s\n' "${latest_release}" "${current_release}" | sort -V | tail -n1)
if [ "${latest_release}" = "${expected_latest_release}" ]; then
  echo "Latest release ${latest_release} is expected, so it doesn't have to be adjusted"
  make_latest="false"
else
  echo "Latest release ${latest_release} should be changed to ${current_release}"
  make_latest="true"
fi

echo "Publishing release ${current_release}"
curl -s -S -f -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${github_api_repo_url}/releases/${release_id}" \
  -d "{\"draft\":false,\"make_latest\":\"${make_latest}\"}"
