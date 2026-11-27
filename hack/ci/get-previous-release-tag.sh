#!/usr/bin/env bash

# Prints the previous release tag for a given release tag.
#
# Usage:
#   get-previous-release-tag.sh <release_tag>
#
# Counts down from (X.Y.Z-1) looking for an existing tag. When patch reaches
# -1, rolls to the next lower minor at patch 0 (not 999 — the previous minor
# series starts at .0). Same for minor rolling to major. The starting point is
# (patch-1) not (patch) because the tag may not exist yet.

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_tools git
require_positional release_tag "$1"

major=$(echo "${release_tag}" | cut -d. -f1)
minor=$(echo "${release_tag}" | cut -d. -f2)
patch=$(echo "${release_tag}" | cut -d. -f3)

all_tags=$(git tag --list | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V)
echo "Found $(echo "${all_tags}" | grep -c . || true) release tag(s) in total" >&2

echo "Finding previous tag for ${release_tag}..." >&2

cur_major=${major}
cur_minor=${minor}
cur_patch=$((patch - 1))
previous=""
while [ -z "${previous}" ] && [ "${cur_major}" -ge 0 ]; do
  while [ -z "${previous}" ] && [ "${cur_minor}" -ge 0 ]; do
    while [ -z "${previous}" ] && [ "${cur_patch}" -ge 0 ]; do
      tag="${cur_major}.${cur_minor}.${cur_patch}"
      if echo "${all_tags}" | grep -qx "${tag}"; then
        previous="${tag}"
      fi
      cur_patch=$((cur_patch - 1))
    done
    cur_minor=$((cur_minor - 1))
    cur_patch=0
  done
  cur_major=$((cur_major - 1))
  cur_minor=0
  cur_patch=0
done

if [ -z "${previous}" ]; then
  echo "ERROR: No previous release tag found" >&2
  exit 1
fi

echo "Previous release tag: ${previous}" >&2
echo "${previous}"
