#!/usr/bin/env bash

# Script generates changelog for the release

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

release_tag=$1
changelog_output_file=$2

repository="${REPOSITORY:-kyma-project/api-gateway}"
github_api_repo_url="https://api.github.com/repos/${repository}"

echo "Create changelog: repository: ${repository}, release tag: ${release_tag}, output file: ${changelog_output_file}"

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
echo -e "**Full changelog:** https://github.com/${repository}/compare/${latest_tag}...${release_tag}" > "${changelog_output_file}"
