#!/usr/bin/env bash

# Description: This script runs given e2e tests on a k3d cluster
# It installs istio and api gateway and then runs make test targets provided via commandline arguments to that script
# It requires the following env variables:
# - IMG - API gateway image to be deployed (by make deploy)
# - K3D_CONFIGURATION - configuration preset, used to load configurations/${K3D_CONFIGURATION}/vars.sh

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_positional make_target "$1"

require_vars K3D_CONFIGURATION IMG

load_configuration "${K3D_CONFIGURATION}"

export TEST_DOMAIN="local.kyma.dev"
export IS_GARDENER=false

# Add pwd to path to be able to use binaries downloaded in scripts
export PATH="${PATH}:${PWD}"

start_group "Creating kyma-system namespace"
make create-namespace
end_group

start_group "Creating kyma-provisioning-info configmap"
make create-provisioning-info DUAL_STACK_ENABLED=false
end_group

start_group "Installing istio"
make install-istio
end_group

start_group "Deploying api-gateway, image: ${IMG}"
make deploy
end_group

start_group "Executing tests: ${make_target}"
make "${make_target}"
end_group
