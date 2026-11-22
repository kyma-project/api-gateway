#!/usr/bin/env bash

# Description: Downloads the k3d binary for the current platform, verifies its checksum
# against the value pinned in vars.sh, and installs it to hack/ci/k3d/bin/k3d (gitignored).
# Run once on a new machine before using provision.sh.
# Environment variables:
# - K3D_CONFIGURATION - configuration preset, used to load configurations/${K3D_CONFIGURATION}/vars.sh

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars K3D_CONFIGURATION
load_configuration "${K3D_CONFIGURATION}"
setup_local_bin

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "${arch}" in
    x86_64)        arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) >&2 echo "Unsupported architecture: ${arch}"; exit 1 ;;
esac

binary="k3d-${os}-${arch}"
checksum_var="K3D_CHECKSUM_$(echo "${os}_${arch}" | tr '[:lower:]' '[:upper:]')"
expected_checksum="${!checksum_var}"
if [ -z "${expected_checksum}" ]; then
    >&2 echo "No pinned checksum found for ${os}/${arch} (expected variable: ${checksum_var})"
    exit 1
fi

base_url="https://github.com/k3d-io/k3d/releases/download/${K3D_VERSION}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

echo "Downloading k3d ${K3D_VERSION} (${os}/${arch})..."
curl -fsSL "${base_url}/${binary}" -o "${tmp_dir}/${binary}"

echo "Verifying checksum..."
# shasum is available on both macOS and Linux; -a 256 selects SHA-256
echo "${expected_checksum}  ${tmp_dir}/${binary}" | shasum -a 256 --check --quiet

install -m 0755 "${tmp_dir}/${binary}" "${script_dir}/bin/k3d"

# Make the bin directory available to subsequent steps in GitHub Actions
if [ -n "${GITHUB_PATH:-}" ]; then
    echo "${script_dir}/bin" >> "${GITHUB_PATH}"
fi

echo "k3d installed successfully: $(k3d version | head -1)"
