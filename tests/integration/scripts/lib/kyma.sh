#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=tests/integration/scripts/lib/log.sh
source "${LIBDIR}/../lib/log.sh"

# kyma::deploy_kyma starts Kyma deployment using new installation method
# Arguments:
# optional:
# s - Kyma source
# d - Kyma workspace directory
# p - execution profile
# u - upgrade (this will not reuse helm values which is already set)
function kyma::deploy_kyma() {
    local OPTIND
    local executionProfile=
    local kymaSource=""
    local kymaSourcesDir=""
    local upgrade=

    log::info "Checking Kyma optional arguments"
    while getopts ":s:p:d:u:" opt; do
        case $opt in
            s)
                kymaSource="$OPTARG"
    						log::info "Kyma Source to install: ${kymaSource}"
                ;;
            p)
                if [ -n "$OPTARG" ]; then
                    executionProfile="$OPTARG"
                    log::info "Execution Profile: ${executionProfile}"
                fi ;;
            d)
                kymaSourcesDir="$OPTARG"
                log::info "Kyma Source Directory: ${kymaSourcesDir}"
                 ;;
            u)
                upgrade="$OPTARG"
                log::info "Kyma upgrade option: ${upgrade}"
                ;;
            \?)
                log::error "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                log::warn "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    log::info "Deploying Kyma"

    if [[ -n "$kymaSource" ]]; then
        kyma deploy --ci --concurrency=8 --profile "$executionProfile" --source="${kymaSource}" --workspace "${kymaSourcesDir}" --verbose
    else
        if [[ -n "$executionProfile" ]]; then
            kyma deploy --ci --concurrency=8 --profile "$executionProfile" --source=local --workspace "${kymaSourcesDir}" --verbose
        else
            kyma deploy --ci --concurrency=8 --source=local --workspace "${kymaSourcesDir}" --verbose
        fi
    fi
}

kyma::install_cli() {
    local settings
    local kyma_version
    settings="$(set +o); set -$-"
    mkdir -p "/tmp/bin"
    export PATH="/tmp/bin:${PATH}"
    os=$(host::os)

    pushd "/tmp/bin" || exit

    log::info "--> Install kyma CLI ${os} locally to /tmp/bin"

    if [[ "${KYMA_MAJOR_VERSION-}" == "1" ]]; then
        curl -sSLo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/1.24.8/kyma_${os}_x86_64.tar.gz"
        tar xvzf kyma.tar.gz
    else
        curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-stable/kyma-${os}?alt=media"
    fi

    chmod +x kyma
    kyma_version=$(kyma version --client)
    log::info "--> Kyma CLI version: ${kyma_version}"
    log::info "OK"
    popd || exit
    eval "${settings}"
}

host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=darwin
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      echo "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}
