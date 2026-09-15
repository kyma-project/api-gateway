#!/usr/bin/env bash

# Prints the reference release version to stdout (the version to upgrade from in upgrade tests).
# Logs are written to stderr.
#
# Required env variables:
#   TARGET_BRANCH: the branch being tested (e.g. main, release-1.2, feat/foo)

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_vars TARGET_BRANCH

release_repository_file="${script_dir}/../../RELEASE_REPOSITORY"
REPO=$(cat "${release_repository_file}")

minor_version_file="${script_dir}/../../MINOR_VERSION"
IFS='.' read -r MAJOR MINOR < "${minor_version_file}"

# Validate release branch consistency
if [[ "${TARGET_BRANCH}" =~ ^release-(.+)$ ]] && [[ "${BASH_REMATCH[1]}" != "${MAJOR}.${MINOR}" ]]; then
  >&2 echo "ERROR: branch '${TARGET_BRANCH}' does not match MINOR_VERSION (${MAJOR}.${MINOR})"
  exit 1
fi

all_releases() {
  gh release list --repo "${REPO}" --json tagName --jq '.[].tagName' \
    | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' \
    | sort -V
}

>&2 echo "Trying to determine reference release for the branch ${TARGET_BRANCH} and repository ${REPO}"

REFERENCE_RELEASE=$(all_releases \
  | awk -F. -v major="${MAJOR}" -v minor="${MINOR}" \
      '$1 < major || ($1 == major && $2 <= minor)' \
  | tail -1)
>&2 echo "Determined reference release: ${REFERENCE_RELEASE}"

if [[ -z "${REFERENCE_RELEASE}" ]]; then
  >&2 echo "ERROR: could not determine reference release for branch '${TARGET_BRANCH}'"
  exit 1
fi

echo "${REFERENCE_RELEASE}"
