#!/usr/bin/env bash

# This script returns the id of the draft release

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

release_tag=$1
release_notes_path=$2
changelog_file_path=$3
release_id_output_file=$4

repository="${REPOSITORY:-kyma-project/api-gateway}"
github_api_repo_url="https://api.github.com/repos/${repository}"

echo "Create draft release: repository: ${repository}, release notes path: ${release_notes_path}, changelog path: ${changelog_file_path}, output file: ${release_id_output_file}"

echo "Preparing release payload"
body=$(cat "${release_notes_path}" <(echo) "${changelog_file_path}")
json_payload=$(jq -n \
  --arg tag_name "${release_tag}" \
  --arg name "${release_tag}" \
  --arg body "${body}" \
  '{
    "tag_name": $tag_name,
    "name": $name,
    "body": $body,
    "draft": true
  }')

echo "Creating release"
curl_response=$(curl -s -S -f -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${github_api_repo_url}/releases" \
  -d "${json_payload}")

echo "Storing release ID in file ${release_id_output_file}"
echo "${curl_response}" | jq -r ".id" > "${release_id_output_file}"
