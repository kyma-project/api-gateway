#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=tests/integration/scripts/lib/log.sh
source "${LIBDIR}"/../lib/log.sh

# utils::check_required_vars checks if all provided variables are initialized
#
# Arguments
# $1 - list of variables
function utils::check_required_vars() {
  log::info "Checks if all provided variables are initialized"
    local discoverUnsetVar=false
    for var in "$@"; do
      if [ -z "${!var}" ] ; then
        log::warn "ERROR: $var is not set"
        discoverUnsetVar=true
      fi
    done
    if [ "${discoverUnsetVar}" = true ] ; then
      exit 1
    fi
}

# utils::save_psp_list generates pod-security-policy list and saves it to json file
#
# Arguments
# $1 - Name of the output json file
function utils::save_psp_list() {
  log::info "generates pod-security-policy list and saves it to json file"
  log::info "json file name: $1"
  if [ -z "$1" ]; then
    echo "File name is empty. Exiting..."
    exit 1
  fi
  local output_path=$1

  # this is false-positive as we need to use single-quotes for jq
  # shellcheck disable=SC2016
  PSP_LIST=$(kubectl get pods --all-namespaces -o json | jq '{ pods: [ .items[] | .metadata.ownerReferences[0].name as $owner | .metadata.annotations."kubernetes.io\/psp" as $psp | { name: .metadata.name, namespace: .metadata.namespace, owner: $owner, psp: $psp} ] | group_by(.name) | map({ name: .[0].name, namespace: .[0].namespace, owner: .[0].owner, psp: .[0].psp }) | sort_by(.psp, .name)}' )
  echo "${PSP_LIST}" > "${output_path}"
}

# utils::describe_nodes call k8s statistics commands and check if oom event was recorded.
function utils::describe_nodes() {
    {
      log::info "calling describe nodes"
      kubectl describe nodes
      log::info "calling top nodes"
      kubectl top nodes
      log::info "calling top pods"
      kubectl top pods --all-namespaces
    } > "${ARTIFACTS}/describe_nodes.txt"
    grep "System OOM encountered" "${ARTIFACTS}/describe_nodes.txt"
    last=$?
    if [[ $last -eq 0 ]]; then
      log::banner "OOM event found"
    fi
}

# utils::oom_get_output download output from debug command pod if exist.
function utils::oom_get_output() {
  if [ ! -e "${ARTIFACTS}/describe_nodes.txt" ]; then
    utils::describe_nodes
  fi
  if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
  log::info "Download OOM events details"
  pods=$(kubectl get pod -l "name=oom-debug" -o=jsonpath='{.items[*].metadata.name}')
  for pod in $pods; do
    kubectl logs "$pod" -c oom-debug > "${ARTIFACTS}/$pod.txt"
  done
  debugFiles=$(ls -1 "${ARTIFACTS}"/oom-debug-*.txt)
  for debugFile in $debugFiles; do
    grep "OOM event received" "$debugFile" > /dev/null
    last=$?
    if [[ $last -eq 0 ]]; then
      log::info "Print OOM events details"
      cat "$debugFile"
    else
      rm "$debugFile"
    fi
  done
  fi
}

# utils::generate_commonName generate common name
# It generates random string and prefix it, with provided arguments
#
# Arguments:
#
# optional:
# n - string to use as a common name prefix
# p - pull request number or commit id to use as a common name prefix
#
# Return:
# utils_generate_commonName_return_commonName - generated common name string
utils::generate_commonName() {

    local OPTIND

    while getopts ":n:p:" opt; do
        case $opt in
            n)
                local namePrefix="$OPTARG" ;;
            p)
                local id="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2; ;;
        esac
    done

    namePrefix=$(echo "$namePrefix" | tr '_' '-')
    namePrefix=${namePrefix#-}

    local randomNameSuffix
    randomNameSuffix=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c6)
    # return value
    # shellcheck disable=SC2034
    utils_generate_commonName_return_commonName=$(echo "$namePrefix$id$randomNameSuffix" | tr "[:upper:]" "[:lower:]" )
}

# check_empty_arg will check if first argument is empty.
# If it's empty it will log second argument as error message and exit with code 1.
# If second argument is empty, generic default log message will be used.
#
# Arguments:
# $1 - argument to check if it's empty
# $2 - log message to use if $1 is empty
function utils::check_empty_arg {
    if [ -z "$2" ]; then
        logMessage="Mandatory argument is empty."
    else
        logMessage="$2"
    fi
    if [ -z "$1" ]; then
        if [ -n "$3" ]; then
            log::error "$logMessage"
        else
            log::error "$logMessage Exiting"
            exit 1
        fi
    fi
}

# utils::generate_vars_for_build generate string values for specific build types
#
# Arguments:
#
# optional:
# b - build type
# p - pull request number, required for build type pr
# s - pull request base SHA, required for build type commit
# n - prowjob name required for other build types
#
# Return variables:
# utils_set_vars_for_build_return_commonName - generated common name
# utils_set_vars_for_build_return_kymaSource - generated kyma source
function utils::generate_vars_for_build {
  log::info "Generate string values for specific build types"

    local OPTIND

    while getopts ":b:p:s:n:" opt; do
        case $opt in
            b)
                local buildType="$OPTARG" ;;
            p)
                local prNumber="$OPTARG" ;;
            s)
                local prBaseSha="$OPTARG" ;;
            n)
                local prowjobName="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
   done

    if [ "$buildType" = "pr" ]; then
        utils::check_empty_arg "$prNumber" "Pull request number not provided."
    fi

    # In case of PR, operate on PR number
    if [[ "$buildType" == "pr" ]]; then
        utils::generate_commonName \
            -n "pr" \
            -p "$prNumber"
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="PR-$prNumber"
    elif [[ "$buildType" == "release" ]]; then
        log::info "Reading release version from VERSION file"
        readonly releaseVersion=$(cat "VERSION")
        log::info "Read release version: $releaseVersion"
        utils::generate_commonName \
            -n "rel"
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="$releaseVersion"
    # Otherwise (master), operate on triggering commit id
    elif [ -n "$prBaseSha" ]; then
        readonly commitID="${prBaseSha::8}"
        utils::generate_commonName \
            -n "commit" \
            -p "$commitID"
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="$commitID"
    elif [ -n "$prowjobName" ]; then
        prowjobName=${prowjobName: -20:20}
        utils::generate_commonName \
            -n "$prowjobName"
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_commonName=${utils_generate_commonName_return_commonName:?}
        # shellcheck disable=SC2034
        utils_generate_vars_for_build_return_kymaSource="main"
    else
        log::error "Build type not known. Set -b parameter to value 'pr' or 'release', or set -s or -n parameter."
    fi
}
