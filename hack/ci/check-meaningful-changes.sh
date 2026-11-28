#!/usr/bin/env bash
# Check whether a release branch has meaningful changes since its last release tag.
# Usage: check-meaningful-changes.sh <branch>
# Must be run from a full checkout with tags fetched.
# Exits: 0 = meaningful changes found, 1 = no meaningful changes (skip), 2 = error

set -euo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional branch "${1:-}"
require_tools git

# Skip if the branch tip is already tagged — the current state has already been released.
branch_tip_tag=""
grep_exit=0
branch_tip_tag=$(git tag --points-at "origin/${branch}" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1) || grep_exit=$?
if [[ "${grep_exit}" -ne 0 && "${grep_exit}" -ne 1 ]]; then
  echo "ERROR: failed to check tags at origin/${branch} (exit ${grep_exit})" >&2
  exit 2
fi
if [[ -n "${branch_tip_tag}" ]]; then
  echo "Branch tip of ${branch} is already tagged as ${branch_tip_tag} — nothing new to release"
  exit 1
fi

last_tag=$(git tag --merged "origin/${branch}" \
  | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' \
  | sort -V \
  | tail -1)

if [[ -z "${last_tag}" ]]; then
  echo "No existing release tag on ${branch} — treating as meaningful (first patch)"
  exit 0
fi

echo "Last tag on ${branch}: ${last_tag}"

changed_files=$(git diff --name-only "${last_tag}...origin/${branch}")
echo "Changed files since ${last_tag}:"
echo "${changed_files}"

meaningful=""
grep_exit=0
meaningful=$(echo "${changed_files}" \
  | grep -v -E '^(docs/|\.github/|OWNERS$|CODEOWNERS$|.*\.md$)') || grep_exit=$?
# grep exits 1 when nothing passes the filter — that means all changes are non-meaningful
if [[ "${grep_exit}" -ne 0 && "${grep_exit}" -ne 1 ]]; then
  echo "ERROR: failed to filter changed files (exit ${grep_exit})" >&2
  exit 2
fi

if [[ -z "${meaningful}" ]]; then
  echo "No meaningful changes since ${last_tag} on ${branch} — skipping"
  exit 1
fi

echo "Meaningful changes found:"
echo "${meaningful}"
exit 0
