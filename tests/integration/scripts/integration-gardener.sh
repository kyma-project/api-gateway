#!/usr/bin/env bash

#Description: Kyma CLI Integration plan on Gardener. This scripts implements a pipeline that consists of many steps. The purpose is to install and test Kyma using the CLI on a real Gardener cluster.
#
#Expected common vars:
# - JOB_TYPE - set up by prow (presubmit, postsubmit, periodic)
# - KYMA_PROJECT_DIR - directory path with Kyma sources to use for installation
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
# - MACHINE_TYPE - (optional) machine type
#
#Please look in each provider script for provider specific requirements

# exit on error, and raise error when variable is not set when used
set -e

export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export API_GATEWAY_SOURCES_DIR="${KYMA_PROJECT_DIR}/api-gateway"

# shellcheck source=tests/integration/scripts/lib/log.sh
source "${API_GATEWAY_SOURCES_DIR}/tests/integration/scripts/lib/log.sh"
# shellcheck source=tests/integration/scripts/lib/utils.sh
source "${API_GATEWAY_SOURCES_DIR}/tests/integration/scripts/lib/utils.sh"
# shellcheck source=tests/integration/scripts/lib/kyma.sh
source "${API_GATEWAY_SOURCES_DIR}/tests/integration/scripts/lib/kyma.sh"
# shellcheck source=tests/integration/scripts/gardener/gardener.sh
source "${API_GATEWAY_SOURCES_DIR}/tests/integration/scripts/gardener/gardener.sh"
# shellcheck source=tests/integration/scripts/integration-tests.sh
source "${API_GATEWAY_SOURCES_DIR}/tests/integration/scripts/integration-tests.sh"

# All provides require these values, each of them may check for additional variables
requiredVars=(
    GARDENER_PROVIDER
    KYMA_PROJECT_DIR
    GARDENER_REGION
    GARDENER_ZONES
    GARDENER_CLUSTER_VERSION
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
)

utils::check_required_vars "${requiredVars[@]}"

if [[ $GARDENER_PROVIDER == "gcp" ]]; then
    # shellcheck source=tests/integration/scripts/gardener/gcp.sh
    source "${API_GATEWAY_SOURCES_DIR}/tests/integration/scripts/gardener/gcp.sh"
else
    log::error "GARDENER_PROVIDER ${GARDENER_PROVIDER} is not yet supported"
    exit 1
fi

# nice cleanup on exit, be it succesful or on fail
trap gardener::cleanup EXIT INT

#Used to detect errors for logging purposes
ERROR_LOGGING_GUARD="true"
export ERROR_LOGGING_GUARD

readonly COMMON_NAME_PREFIX="grd"
utils::generate_commonName -n "${COMMON_NAME_PREFIX}"
COMMON_NAME="${utils_generate_commonName_return_commonName:?}"
export COMMON_NAME

export CLUSTER_NAME="${COMMON_NAME}"

# set KYMA_SOURCE used by kyma deploy
utils::generate_vars_for_build \
    -b "$BUILD_TYPE" \
    -p "$PULL_NUMBER" \
    -s "$PULL_BASE_SHA" \
    -n "$JOB_NAME"
export KYMA_SOURCE=${utils_generate_vars_for_build_return_kymaSource:?}

# checks required vars and initializes gcloud/docker if necessary
gardener::init

# if MACHINE_TYPE is not set then use default one
gardener::set_machine_type

kyma::install_cli

# currently only Azure generates overrides, but this may change in the future
gardener::generate_overrides

export CLEANUP_CLUSTER="true"
gardener::provision_cluster

api-gateway::prepare_components_file_istio_only
api-gateway::prepare_test_env_integration_tests

# generate pod-security-policy list in json
utils::save_psp_list "${ARTIFACTS}/kyma-psp.json"

if [[ "${HIBERNATION_ENABLED}" == "true" ]]; then
    gardener::hibernate_kyma
    sleep 120
    gardener::wake_up_kyma
fi

"${API_GATEWAY_SOURCES_DIR}/tests/integration/scripts/jobguard/run.sh"
api-gateway::launch_integration_tests

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
