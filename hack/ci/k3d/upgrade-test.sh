#!/usr/bin/env bash

# Description: This script runs upgrade tests on a k3d cluster.
# It deploys the reference release, then runs the given make target which
# handles the upgrade using IMG and validates the upgrade behavior.
# It requires the following env variables:
# - REFERENCE_RELEASE - GitHub release to deploy as the baseline before upgrading
# - IMG - image used by the make target to perform and validate the upgrade
# - K3D_CONFIGURATION - configuration preset, used to load configurations/${K3D_CONFIGURATION}/vars.sh

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_positional make_target "$1"

require_vars K3D_CONFIGURATION REFERENCE_RELEASE IMG

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

start_group "Deploy reference release of a module"
echo "Deploying from release: ${REFERENCE_RELEASE}"
make deploy-release RELEASE_VERSION="${REFERENCE_RELEASE}"
end_group

start_group "Executing tests: ${make_target}"
make "${make_target}"
end_group
