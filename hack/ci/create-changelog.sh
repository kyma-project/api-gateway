#!/usr/bin/env bash

# Script generates changelog for the release

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional release_tag "$1"
require_positional changelog_output_file "$2"
require_vars GITHUB_TOKEN

REPOSITORY=$(gh repo view --json nameWithOwner -q .nameWithOwner)
github_api_repo_url="https://api.github.com/repos/${REPOSITORY}"

echo "Create changelog: repository: ${REPOSITORY}, release tag: ${release_tag}, output file: ${changelog_output_file}"

echo "Fetching all release tags"
tags=$(curl -s -S -f -L \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  "${github_api_repo_url}/releases" \
  | jq -r .[].tag_name \
  | sort -V)

echo "Parse release tag ${release_tag}"
major=$(echo "${release_tag}" | cut -d. -f1)
minor=$(echo "${release_tag}" | cut -d. -f2)
patch=$(echo "${release_tag}" | cut -d. -f3)
echo "Major: ${major}, minor: ${minor}, patch: ${patch}"

echo "Determine previous version for changelog"
latest_tag=""
if [ "${patch}" -ne 0 ]; then
  latest_tag=$(echo "${tags}" | grep -E "^${major}\.${minor}\." | tail -1)
elif [ "${minor}" -ne 0 ]; then
  prev_minor=$((minor - 1))
  latest_tag=$(echo "${tags}" | grep -E "^${major}\.${prev_minor}\." | head -n 1)
else
  prev_major=$((major - 1))
  latest_tag=$(echo "${tags}" | grep -E "^${prev_major}\." | head -n 1)
fi
echo "Previous version: ${latest_tag}"

echo "Storing changelog in ${changelog_output_file}"
echo -e "**Full changelog:** https://github.com/${REPOSITORY}/compare/${latest_tag}...${release_tag}" > "${changelog_output_file}"
