#!/usr/bin/env bash

# Compute the next patch version for this repo.
# Reads major.minor from MINOR_VERSION in the repo root.
# Prints: X.Y.(N+1) based on the highest git tag, or X.Y.0 if no tags exist.

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_tools git

repo_root="${script_dir}/../.."
IFS='.' read -r major minor < "${repo_root}/MINOR_VERSION"

echo "Computing next patch for ${major}.${minor}.* from git tags..." >&2

highest=""
grep_exit=0
highest=$(git tag --merged HEAD | grep -E "^${major}\.${minor}\.[0-9]+$" | sort -V | tail -1) || grep_exit=$?
if [[ "${grep_exit}" -ne 0 && "${grep_exit}" -ne 1 ]]; then
  echo "ERROR: failed to query git tags (exit ${grep_exit})" >&2
  exit 1
fi

if [ -z "${highest}" ]; then
  echo "No existing tags for ${major}.${minor}.* — starting at ${major}.${minor}.0" >&2
  echo "${major}.${minor}.0"
else
  current_patch=$(echo "${highest}" | cut -d. -f3)
  next_patch=$(( current_patch + 1 ))
  echo "Latest tag: ${highest}, next: ${major}.${minor}.${next_patch}" >&2
  echo "${major}.${minor}.${next_patch}"
fi
