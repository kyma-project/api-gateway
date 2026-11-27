#!/usr/bin/env bash

# Creates or updates a draft GitHub release with assembled release notes.
#
# Usage: create-draft-release.sh <release_tag> <release_id_output_file>

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional release_tag "$1"
require_positional release_id_output_file "$2"
require_vars GITHUB_TOKEN
require_tools gh git jq curl envsubst

repository=$(gh repo view --json nameWithOwner -q .nameWithOwner)
github_api_repo_url="https://api.github.com/repos/${repository}"
current_branch=$(git rev-parse --abbrev-ref HEAD)

echo "Create draft release: repository: ${repository}, branch: ${current_branch}, tag: ${release_tag}, output file: ${release_id_output_file}"

echo "Computing changelog URL"
export CHANGELOG_URL
previous_tag=$("${script_dir}/get-previous-release-tag.sh" "${release_tag}")
echo "Previous tag: ${previous_tag}"
CHANGELOG_URL="https://github.com/${repository}/compare/${previous_tag}...${release_tag}"

echo "Assembling release notes from PR descriptions"
release_notes_file=$(mktemp)
"${script_dir}/assemble-release-notes.sh" "${release_notes_file}" "${previous_tag}" HEAD
export RELEASE_NOTES
RELEASE_NOTES=$(cat "${release_notes_file}")

echo "Rendering release body"
body_file=$(mktemp)
envsubst '$RELEASE_NOTES $CHANGELOG_URL' \
  < "${script_dir}/release-body.md.tmpl" > "${body_file}"
body=$(cat "${body_file}")

echo "Looking for existing draft release with name ${release_tag}..."
existing_releases=$(gh api "repos/${repository}/releases")

# GitHub renames tag_name to "untagged-<hash>" for draft releases where the tag
# doesn't exist yet, so we match on .name (the release title) instead.
published_id=$(echo "${existing_releases}" \
  | jq -r ".[] | select(.name == \"${release_tag}\" and .draft == false) | .id")
if [ -n "${published_id}" ]; then
  echo "ERROR: a published release named ${release_tag} already exists (id: ${published_id}) — refusing to overwrite" >&2
  exit 1
fi

existing_ids=$(echo "${existing_releases}" \
  | jq -r ".[] | select(.name == \"${release_tag}\" and .draft == true) | .id")
existing_id=$(echo "${existing_ids}" | head -1)

if [ "$(echo "${existing_ids}" | wc -l)" -gt 1 ]; then
  echo "WARNING: multiple draft releases named ${release_tag} found — updating the first (id: ${existing_id})"
fi

if [ -n "${existing_id}" ]; then
  echo "Found existing draft release (id: ${existing_id}), updating (tag_name: ${release_tag})..."
  curl -s -S -f -L \
    -X PATCH \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "${github_api_repo_url}/releases/${existing_id}" \
    -d "$(jq -n --arg tag_name "${release_tag}" --arg body "${body}" --arg branch "${current_branch}" \
      '{"tag_name": $tag_name, "body": $body, "target_commitish": $branch}')" > /dev/null
  echo "Updated draft release (id: ${existing_id})"
  release_id="${existing_id}"
else
  echo "No existing draft release found, creating new draft for ${release_tag}..."
  json_payload=$(jq -n \
    --arg tag_name "${release_tag}" \
    --arg name "${release_tag}" \
    --arg body "${body}" \
    --arg branch "${current_branch}" \
    '{"tag_name": $tag_name, "name": $name, "body": $body, "draft": true, "target_commitish": $branch}')
  release_id=$(curl -s -S -f -L \
    -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "${github_api_repo_url}/releases" \
    -d "${json_payload}" | jq -r '.id')
  echo "Created draft release (id: ${release_id})"
fi

echo "Storing release ID in file ${release_id_output_file}"
echo "${release_id}" > "${release_id_output_file}"
