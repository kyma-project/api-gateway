#!/usr/bin/env bash

# Verify that every image listed in external-images.yaml exists on Docker Hub.
# Exits non-zero if any image is missing or if an unexpected HTTP error occurs.

set -euo pipefail

script_dir="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_tools curl yq

IMAGES_FILE="${script_dir}/../../external-images.yaml"

main() {
  local count i src code body_file failed

  body_file="$(mktemp)"
  trap "rm -f '${body_file}'" EXIT

  failed=false
  count="$(yq '.images | length' "${IMAGES_FILE}")"
  for i in $(seq 0 $((count - 1))); do
    src="$(yq -r ".images[${i}].source" "${IMAGES_FILE}")"
    image_without_tag="${src%%:*}"
    tag="${src#*:}"
    echo "Checking Docker Hub: ${src}"
    code="$(curl -sSL -o "${body_file}" -w '%{http_code}' \
      "https://hub.docker.com/v2/repositories/${image_without_tag}/tags/${tag}" \
      2>/dev/null)" || true
    if [ "${code}" = "200" ]; then
      echo "Available: ${src}"
    elif [ "${code}" = "404" ]; then
      echo "NOT FOUND on Docker Hub: ${src}" >&2
      failed=true
    else
      echo "Unexpected HTTP ${code} for ${src}:" >&2
      cat "${body_file}" >&2
      exit 1
    fi
  done

  if [ "${failed}" = "true" ]; then
    echo "One or more images not found on Docker Hub" >&2
    exit 1
  fi

  echo "All external images are available on Docker Hub"
}

main
