#!/usr/bin/env bash

# Generates and publishes release assets to a GitHub draft release

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional image_name "$1"
require_positional release_tag "$2"
require_positional release_id "$3"
require_vars GITHUB_TOKEN

REPOSITORY=$(gh repo view --json nameWithOwner -q .nameWithOwner)
github_upload_repo_url="https://uploads.github.com/repos/${REPOSITORY}"

echo "Publish assets: repository: ${REPOSITORY}, image name: ${image_name}, release tag: ${release_tag}, release ID: ${release_id}"

echo "Generate manifests"
IMG="${image_name}:${release_tag}" VERSION="${release_tag}" make generate-manifests

echo "Publish manager deployment"
manager_yaml_asset_name="api-gateway-manager.yaml"
manager_yaml_asset_path="api-gateway-manager.yaml"
curl -s -S -f -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @"${manager_yaml_asset_path}" \
  "${github_upload_repo_url}/releases/${release_id}/assets?name=${manager_yaml_asset_name}"

echo "Publish default CR"
default_cr_asset_name="apigateway-default-cr.yaml"
default_cr_path="config/samples/operator_v1alpha1_apigateway.yaml"
curl -s -S -f -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @"${default_cr_path}" \
  "${github_upload_repo_url}/releases/${release_id}/assets?name=${default_cr_asset_name}"
