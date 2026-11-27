#!/usr/bin/env bash

# Publishes a draft GitHub release, marking it as latest if it is the highest version

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional release_id "$1"
require_positional release_tag "$2"
require_vars GITHUB_TOKEN

REPOSITORY=$(gh repo view --json nameWithOwner -q .nameWithOwner)
github_api_repo_url="https://api.github.com/repos/${REPOSITORY}"

echo "Publish release: repository: ${REPOSITORY}, release ID: ${release_id}, tag: ${release_tag}"

echo "Getting latest release"
# /releases/latest always returns a published release, so .name is a real version here.
latest_release=$(curl -s -S -f -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${github_api_repo_url}/releases/latest" | \
  jq -r '.name')
echo "Latest release = ${latest_release}"

expected_latest_release=$(printf '%s\n' "${latest_release}" "${release_tag}" | sort -V | tail -n1)
if [ "${latest_release}" = "${expected_latest_release}" ]; then
  echo "Latest release ${latest_release} is already the highest, not marking ${release_tag} as latest"
  make_latest="false"
else
  echo "Marking ${release_tag} as latest (was ${latest_release})"
  make_latest="true"
fi

echo "Publishing release ${release_tag}"
curl -s -S -f -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${github_api_repo_url}/releases/${release_id}" \
  -d "{\"draft\":false,\"make_latest\":\"${make_latest}\"}"
