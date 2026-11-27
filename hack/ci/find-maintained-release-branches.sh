#!/usr/bin/env bash
# Prints the maintained release-X.Y branches for this repo (semver sort).
# The number of branches is controlled by MAINTAINED_RELEASE_BRANCHES_COUNT env var.
# Requires: gh CLI, authenticated via gh auth login (local) or GITHUB_TOKEN (CI)

set -euo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_vars MAINTAINED_RELEASE_BRANCHES_COUNT
require_tools gh jq

if ! [[ "${MAINTAINED_RELEASE_BRANCHES_COUNT}" =~ ^[0-9]+$ ]]; then
  echo "ERROR: MAINTAINED_RELEASE_BRANCHES_COUNT must be a number, got: '${MAINTAINED_RELEASE_BRANCHES_COUNT}'" >&2
  exit 1
fi

repository=$(gh repo view --json nameWithOwner -q .nameWithOwner)

gh api --paginate "repos/${repository}/branches" \
  | jq -r '.[].name' \
  | grep -E '^release-[0-9]+\.[0-9]+$' \
  | sort -t'-' -k2 -V \
  | tail -"${MAINTAINED_RELEASE_BRANCHES_COUNT}"
