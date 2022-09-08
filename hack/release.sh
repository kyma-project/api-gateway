#!/usr/bin/env bash

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly API_GATEWAY_DIR="/home/prow/go/src/github.com/kyma-project/api-gateway"

main() {
    git remote add origin git@github.com:kyma-project/api-gateway.git
    # release api-gateway with goreleaser
    curl -sL https://git.io/goreleaser | VERSION=v1.11.2 bash -s --
}

main
