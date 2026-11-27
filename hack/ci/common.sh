#!/usr/bin/env bash

# Sourced by CI scripts. Provides shared utilities.

# require_bash <major>
# Exits if the current shell is not Bash or is older than the required major version.
require_bash() {
  local required_major="$1"
  if [ -z "${BASH_VERSION:-}" ]; then
    echo "This script requires Bash ${required_major}.0+ (current shell: $(ps -p $$ -o comm=))" >&2
    exit 1
  fi
  if (( BASH_VERSINFO[0] < required_major )); then
    echo "Bash ${required_major}.0+ is required (found ${BASH_VERSION})" >&2
    exit 1
  fi
}

# load_configuration <config_name>
# Sources configurations/<config_name>/vars.sh relative to the calling script's directory,
# auto-exporting all variables defined in that file.
load_configuration() {
  local config_name="$1"
  local caller_dir
  caller_dir="$(dirname "$(readlink -f "${BASH_SOURCE[1]}")")"
  local preset_vars="${caller_dir}/configurations/${config_name}/vars.sh"
  if [ ! -f "${preset_vars}" ]; then
    >&2 echo "Configuration file '${preset_vars}' not found"
    exit 2
  fi
  set -a
  # shellcheck source=/dev/null
  source "${preset_vars}"
  set +a
}

# require_vars VAR1 VAR2 ...
# Verifies all named variables are non-empty. Prints all missing ones before exiting.
require_vars() {
  local missing=false
  for var in "$@"; do
    if [ -z "${!var:-}" ]; then
      >&2 echo "Environment variable ${var} is required but not set"
      missing=true
    fi
  done
  if [ "${missing}" = true ]; then
    exit 2
  fi
}

# require_files VAR1 VAR2 ...
# Verifies all named variables point to existing files. Prints all missing ones before exiting.
require_files() {
  local missing=false
  for var in "$@"; do
    local path="${!var}"
    if [ ! -f "${path}" ]; then
      >&2 echo "File '${path}' (${var}) required but not found"
      missing=true
    fi
  done
  if [ "${missing}" = true ]; then
    exit 2
  fi
}

# require_tools TOOL1 TOOL2 ...
# Verifies all named tools are available in PATH. Prints all missing ones before exiting.
require_tools() {
  local missing=false
  for tool in "$@"; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
      >&2 echo "Required tool '${tool}' not found in PATH"
      missing=true
    fi
  done
  if [ "${missing}" = true ]; then
    exit 2
  fi
}

# start_group <title> / end_group
# Emits GitHub Actions log group markers when running in GitHub Actions.
start_group() { [ "${GITHUB_ACTIONS}" = "true" ] && echo "::group::$1" || echo "$1"; }
end_group()   { [ "${GITHUB_ACTIONS}" = "true" ] && echo "::endgroup::" || true; }

# log <message>
# Prints a message with the current log indent level (managed by increase/decrease_log_indent).
log_indent=0
increase_log_indent() { ((log_indent++)) || true; }
decrease_log_indent() { ((log_indent--)) || true; }
log() { printf '%*s%s\n' "$((log_indent * 2))" "" "$*"; }

# require_positional <var_name> <value>
# Assigns <value> to <var_name> and exits if it is empty.
require_positional() {
  local var_name="$1"
  local value="$2"
  if [ -z "${value}" ]; then
    >&2 echo "${var_name} is required as positional argument"
    exit 3
  fi
  printf -v "${var_name}" '%s' "${value}"
}

# setup_local_bin
# Prepends <caller_script_dir>/bin to PATH, creating the directory if needed.
# Scripts that download binaries should install them there.
setup_local_bin() {
  local caller_dir
  caller_dir="$(dirname "$(readlink -f "${BASH_SOURCE[1]}")")"
  local bin_dir="${caller_dir}/bin"
  mkdir -p "${bin_dir}"
  export PATH="${bin_dir}:${PATH}"
}

# check_envsubst_vars <template_file>
# Reads all $VAR / ${VAR} references from the template and verifies they are set.
check_envsubst_vars() {
  local template_file="$1"
  local vars
  # envsubst --variables reads the shell-format string from stdin
  vars=$(envsubst --variables "$(cat "${template_file}")")
  # shellcheck disable=SC2086
  require_vars ${vars}
}
