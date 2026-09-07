#!/usr/bin/env bash

# Publishes a draft GitHub release, marking it as latest if it is the highest version

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional release_id "$1"
require_vars GITHUB_TOKEN

REPOSITORY=$(gh repo view --json nameWithOwner -q .nameWithOwner)
github_api_repo_url="https://api.github.com/repos/${REPOSITORY}"

echo "Publish release: repository: ${REPOSITORY}, release ID: ${release_id}"

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
