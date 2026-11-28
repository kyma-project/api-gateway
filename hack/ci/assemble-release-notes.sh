#!/usr/bin/env bash

# Assemble release notes by walking git log since the given tag and extracting
# ```release-notes-* blocks from merged PR descriptions.
#
# Usage:
#   assemble-release-notes.sh <output_file> <since_tag> <until>
#
# <output_file> file to write the release notes to
# <since_tag>   the tag/commit to start from (exclusive)
# <until>       commit/tag to end at (inclusive), use HEAD for latest
#
# Section titles are derived from block names by stripping the "release-notes-"
# prefix, replacing hyphens with spaces, and capitalizing each word.
# Example: release-notes-new-features -> New Features
#
# Dependabot PRs (detected by "dependencies" label) are auto-added to the
# "Dependency Updates" section even without a release-notes block.
#
# Requires: gh, git, jq, GITHUB_TOKEN in env.

set -eo pipefail
script_dir="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_bash 4

# --- block extraction ---

# Derives a section title from a block name.
# release-notes-new-features -> New Features
section_title_from_block_name() {
  local name="$1"
  local stripped="${name#release-notes-}"
  local spaced="${stripped//-/ }"
  echo "${spaced}" | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) substr($i,2); print}'
}

# Returns all release-notes-* block names found in a PR body.
extract_block_names() {
  local body="$1"
  echo "${body}" | tr -d '\r' | grep -oE '^```release-notes-[a-z][a-z-]*' | cut -c4- || true
}

# Extracts the content of a named release-notes-* block from a PR body.
extract_block_content() {
  local body="$1" block_name="$2"
  echo "${body}" | tr -d '\r' | awk "/^\`\`\`${block_name}/{found=1;next} found && /^\`\`\`/{found=0;next} found{print}"
}

# --- section accumulation ---

# sections maps section_name -> accumulated bullet entries
declare -A sections
declare -a section_order

# Registers a section in the output order without adding content.
register_section() {
  local section="$1"
  if [ -z "${sections[${section}]+x}" ]; then
    section_order+=("${section}")
    sections["${section}"]=""
  fi
}

# Formats a multi-line block as a markdown bullet: first line prefixed with "- ",
# subsequent lines indented with two spaces.
format_bullet() {
  local first=1
  while IFS= read -r line; do
    if [ "${first}" -eq 1 ]; then
      echo "- ${line}"
      first=0
    else
      echo "  ${line}"
    fi
  done <<< "$1"
}

# Appends a PR entry to a section. Registers the section even if block_content is empty.
append_to_section() {
  local section="$1" block_content="$2" pr_link="$3"
  register_section "${section}"
  if [ -z "${block_content// }" ]; then return; fi
  local entry
  entry="$(format_bullet "${block_content}")${pr_link:+ ${pr_link}}"
  if [ -n "${sections[${section}]}" ]; then
    sections["${section}"]=$(printf '%s\n%s' "${sections[${section}]}" "${entry}")
  else
    sections["${section}"]="${entry}"
  fi
}

# --- output rendering ---

render_output() {
  local output="" separator=""
  for section in "${section_order[@]}"; do
    if [ "${section}" = "Dependency Updates" ]; then continue; fi  # appended last below
    printf -v block '%s## %s\n\n%s\n\n' "${separator}" "${section}" "${sections[${section}]}"
    output+="${block}"
    separator=$'\n'
  done
  if [ -n "${sections[Dependency Updates]+x}" ]; then
    printf -v block '%s## %s\n\n%s\n\n' "${separator}" "Dependency Updates" "${sections[Dependency Updates]}"
    output+="${block}"
  fi
  printf '%s' "${output}"
}

# --- main ---

require_positional output_file "$1"
require_positional since_tag "$2"
require_positional until_ref "$3"

require_vars GITHUB_TOKEN
require_tools gh git jq

log "Assembling release notes since ${since_tag} until ${until_ref}..."

git rev-parse --verify "${since_tag}" > /dev/null
since_exit=$?
if [[ "${since_exit}" -ne 0 ]]; then
  echo "ERROR: since ref '${since_tag}' not found in git history" >&2
  exit 1
fi

git rev-parse --verify "${until_ref}" > /dev/null
until_exit=$?
if [[ "${until_exit}" -ne 0 ]]; then
  echo "ERROR: until ref '${until_ref}' not found in git history" >&2
  exit 1
fi

repository=$(gh repo view --json nameWithOwner -q .nameWithOwner)
log "Repository: ${repository}"

declare -A seen_prs

while IFS= read -r log_line; do
  log "Processing commit: ${log_line}"
  increase_log_indent

  pr_number=""
  if [[ "${log_line}" =~ \(#([0-9]+)\)$ ]]; then
    pr_number="${BASH_REMATCH[1]}"
  elif [[ "${log_line}" =~ pull\ request\ #([0-9]+) ]]; then
    pr_number="${BASH_REMATCH[1]}"
  fi

  if [ -z "${pr_number}" ]; then
    log "No PR number found — skipping"
    decrease_log_indent; continue
  fi

  if [ -n "${seen_prs[${pr_number}]+x}" ]; then
    log "PR #${pr_number} already processed — skipping"
    decrease_log_indent; continue
  fi
  seen_prs["${pr_number}"]=1

  log "Fetching PR #${pr_number}..."
  pr_json=""
  if pr_json=$(gh pr view "${pr_number}" --json title,body,labels 2>/dev/null); then
    log "Found PR #${pr_number}"
  else
    log "Warning: could not fetch PR #${pr_number}, skipping"
    decrease_log_indent; continue
  fi

  pr_title=$(echo "${pr_json}" | jq -r '.title')
  pr_body=$(echo "${pr_json}"  | jq -r '.body // ""')
  pr_labels=$(echo "${pr_json}" | jq -r '[.labels[].name] | join(",")')
  log "Title: ${pr_title}"
  log "Labels: ${pr_labels:-none}"

  pr_link="[#${pr_number}](https://github.com/${repository}/pull/${pr_number})"

  block_names=$(extract_block_names "${pr_body}")
  for block_name in ${block_names}; do
    section=$(section_title_from_block_name "${block_name}")
    log "Found block '${block_name}' -> section '${section}'"
    content=$(extract_block_content "${pr_body}" "${block_name}")
    log "Routing to '${section}'"
    append_to_section "${section}" "${content}" "${pr_link}"
  done

  if [ -z "${block_names}" ]; then
    if echo "${pr_labels}" | grep -qE "(^|,)dependencies(,|$)"; then
      append_to_section "Dependency Updates" "${pr_title}" "${pr_link}"
      log "Routing to Dependency Updates (auto-generated from dependencies label)"
    else
      log "Skipping: no release-notes-* blocks and no dependencies label"
    fi
  fi

  decrease_log_indent
done < <(git log --oneline "${since_tag}..${until_ref}")

if [ "${#section_order[@]}" -eq 0 ]; then
  log "No release notes found — using generic fallback"
  append_to_section "Dependency Updates" "This version includes dependency updates and security patches." ""
fi

output="$(render_output)"
printf '%s' "${output}" > "${output_file}"
log "Release notes written to ${output_file}:"
echo "${output}"
