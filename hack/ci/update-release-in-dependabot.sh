#!/usr/bin/env bash

# Updates the release branch in dependabot.yml to match MINOR_VERSION file.

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

minor_version_file="${script_dir}/../../MINOR_VERSION"
if [ ! -f "${minor_version_file}" ]; then
  >&2 echo "ERROR: MINOR_VERSION file not found at ${minor_version_file}"
  exit 1
fi

minor_version=$(cat "${minor_version_file}")
dependabot_file="${script_dir}/../../.github/dependabot.yml"

echo "Updating dependabot.yml release branch to release-${minor_version}"
sed -i'' "s|release-[0-9]*\.[0-9]*|release-${minor_version}|g" "${dependabot_file}"
