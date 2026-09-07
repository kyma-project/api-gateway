#!/usr/bin/env bash

# Creates a draft GitHub release and writes the release ID to a file

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional release_tag "$1"
require_positional release_notes_path "$2"
require_positional changelog_file_path "$3"
require_positional release_id_output_file "$4"
require_vars GITHUB_TOKEN

REPOSITORY=$(gh repo view --json nameWithOwner -q .nameWithOwner)
github_api_repo_url="https://api.github.com/repos/${REPOSITORY}"

echo "Create draft release: repository: ${REPOSITORY}, release notes path: ${release_notes_path}, changelog path: ${changelog_file_path}, output file: ${release_id_output_file}"

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
