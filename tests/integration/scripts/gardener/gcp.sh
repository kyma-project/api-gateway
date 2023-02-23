#!/usr/bin/env bash

#Permissions: In order to run this script you need to use a service account with permissions equivalent to the following GCP roles:
# - Compute Admin
# - Service Account User
# - Service Account Admin
# - Service Account Token Creator
# - Make sure the service account is enabled for the Google Identity and Access Management API.

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

# shellcheck source=tests/integration/scripts/lib/log.sh
source "${LIBDIR}"/log.sh
# shellcheck source=tests/integration/scripts/lib/kyma.sh
source "${LIBDIR}"/kyma.sh
# shellcheck source=tests/integration/scripts/lib/utils.sh
source "${LIBDIR}"/utils.sh
# shellcheck source=tests/integration/scripts/gardener/gardener.sh
source "${LIBDIR}"/../gardener/gardener.sh

#!Put cleanup code in this function! Function is executed at exit from the script and on interuption.
gardener::cleanup() {
    #!!! Must be at the beginning of this function !!!
    EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        log::error "AN ERROR OCCURED! Take a look at preceding log entries."
    fi

    #Turn off exit-on-error so that next step is executed even if previous one fails.
    set +e

    # describe nodes to file in artifacts directory
    utils::describe_nodes

    if [ "${DEBUG_COMMANDO_OOM}" = "true" ]; then
      # copy output from debug container to artifacts directory
      utils::oom_get_output
    fi

    if  [[ "${CLEANUP_CLUSTER}" == "true" ]] ; then
        log::info "Deprovision cluster: \"${CLUSTER_NAME}\""
        gardener::deprovision_cluster \
            -p "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
            -c "${CLUSTER_NAME}" \
            -f "${GARDENER_KYMA_PROW_KUBECONFIG}"
    fi

    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    log::info "Job is finished ${MSG}"
    set -e

    exit "${EXIT_STATUS}"
}


gardener::init() {
    requiredVars=(
        KYMA_PROJECT_DIR
        GARDENER_REGION
        GARDENER_ZONES
        GARDENER_KYMA_PROW_KUBECONFIG
        GARDENER_KYMA_PROW_PROJECT_NAME
        GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
        KYMA_SOURCE
    )

    utils::check_required_vars "${requiredVars[@]}"
}

gardener::set_machine_type() {
    if [ -z "$MACHINE_TYPE" ]; then
        export MACHINE_TYPE="n1-standard-4"
    fi
}

gardener::generate_overrides() {
    # currently only Azure generates anything in this function
    return
}


gardener::provision_cluster() {
    log::info "Provision cluster: \"${CLUSTER_NAME}\""
    if [ "${#CLUSTER_NAME}" -gt 9 ]; then
        log::error "Provided cluster name is too long"
        return 1
    fi

    CLEANUP_CLUSTER="true"
      # enable trap to catch kyma provision failures
      trap gardener::reprovision_cluster ERR
      # decreasing attempts to 2 because we will try to create new cluster from scratch on exit code other than 0
      kyma provision gardener gcp \
        --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" \
        --name "${CLUSTER_NAME}" \
        --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
        --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" \
        -z "${GARDENER_ZONES}" \
        -t "${MACHINE_TYPE}" \
        --scaler-max 4 \
        --scaler-min 2 \
        --kube-version="${GARDENER_CLUSTER_VERSION}" \
        --attempts 1 \
        --verbose
    # trap cleanup we want other errors fail pipeline immediately
    trap - ERR
}

gardener::deploy_kyma() {
    kyma deploy --ci --timeout 90m "$@"
}

gardener::hibernate_kyma() {
    return
}

gardener::wake_up_kyma() {
    return
}
