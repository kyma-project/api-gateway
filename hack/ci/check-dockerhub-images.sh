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
    if [[ "${src}" == *@sha256:* ]]; then
      image_without_tag="${src%%@*}"
      digest="${src#*@}"
      tag="$(yq -r ".images[${i}].tag" "${IMAGES_FILE}")"
    else
      image_without_tag="${src%%:*}"
      tag="${src#*:}"
      digest=""
    fi

    # Official Docker Hub images (no namespace) live under the library/ namespace in the API
    if [[ "${image_without_tag}" != */* ]]; then
      image_without_tag="library/${image_without_tag}"
    fi

    echo "Checking Docker Hub tag: ${image_without_tag}:${tag}"
    code="$(curl -sSL -o "${body_file}" -w '%{http_code}' \
      "https://hub.docker.com/v2/repositories/${image_without_tag}/tags/${tag}" \
      2>/dev/null)" || true
    if [ "${code}" = "200" ]; then
      echo "Available: ${image_without_tag}:${tag}"
    elif [ "${code}" = "404" ]; then
      echo "NOT FOUND on Docker Hub: ${image_without_tag}:${tag}" >&2
      failed=true
    else
      echo "Unexpected HTTP ${code} for ${image_without_tag}:${tag}:" >&2
      cat "${body_file}" >&2
      exit 1
    fi

    if [ -n "${digest}" ]; then
      echo "Checking Docker Hub digest: ${image_without_tag}@${digest}"
      local token
      token="$(curl -sSL "https://auth.docker.io/token?service=registry.docker.io&scope=repository:${image_without_tag}:pull" | yq -r '.token')"
      code="$(curl -sSL -o "${body_file}" -w '%{http_code}' \
        -H "Authorization: Bearer ${token}" \
        -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
        "https://index.docker.io/v2/${image_without_tag}/manifests/${digest}" \
        2>/dev/null)" || true
      if [ "${code}" = "200" ]; then
        echo "Available: ${image_without_tag}@${digest}"
      elif [ "${code}" = "404" ]; then
        echo "NOT FOUND on Docker Hub: ${image_without_tag}@${digest}" >&2
        failed=true
      else
        echo "Unexpected HTTP ${code} for ${image_without_tag}@${digest}:" >&2
        cat "${body_file}" >&2
        exit 1
      fi
    fi
  done

  if [ "${failed}" = "true" ]; then
    echo "One or more images not found on Docker Hub" >&2
    exit 1
  fi

  echo "All external images are available on Docker Hub"
}

main
