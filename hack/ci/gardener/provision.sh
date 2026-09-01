#!/usr/bin/env bash

# Description: This script provisions a Gardener cluster
# It requires the following env variables:
# - CLUSTER_NAME - name of the cluster to be created
# - CLUSTER_KUBECONFIG - target path where the kubeconfig of the newly created cluster is stored
# - GARDENER_KUBECONFIG - Gardener kubeconfig path
# - GARDENER_PROJECT_NAME - name of the Gardener project
# - GARDENER_CONFIGURATION - provisioning preset, selects configurations/${GARDENER_CONFIGURATION}/
# All other variables (provider, region, machine type, k8s version, ...) are
# loaded from configurations/${GARDENER_CONFIGURATION}/vars.sh and the shoot template from
# configurations/${GARDENER_CONFIGURATION}/shoot.yaml

set -eo pipefail
echo "::group::Provision Gardener cluster"
script_dir="$(dirname "$(readlink -f "$0")")"

function check_required_vars() {
  local requiredVarMissing=false
  for var in "$@"; do
    if [ -z "${!var}" ]; then
      >&2 echo "Environment variable ${var} is required but not set"
      requiredVarMissing=true
    fi
  done
  if [ "${requiredVarMissing}" = true ] ; then
    echo "::endgroup::"
    exit 2
  fi
}

function check_required_files() {
  local requiredFileMissing=false
  for file in "$@"; do
    path=$(eval echo "\$$file")
    if [ ! -f "${path}" ]; then
        >&2 echo "File '${path}' required but not found"
        requiredFileMissing=true
    fi
  done
  if [ "${requiredFileMissing}" = true ] ; then
    echo "::endgroup::"
    exit 2
  fi
}

check_required_vars GARDENER_CONFIGURATION
preset_dir="${script_dir}/configurations/${GARDENER_CONFIGURATION}"
if [ ! -f "${preset_dir}/vars.sh" ]; then
    >&2 echo "File '${preset_dir}/vars.sh' required but not found"
    echo "::endgroup::"
    exit 2
fi
set -a # autoexport variables in the sourced file
source "${preset_dir}/vars.sh"
set +a

requiredVars=(
    CLUSTER_NAME
    CLUSTER_KUBECONFIG
    GARDENER_PROVIDER
    GARDENER_IP_STACK
    GARDENER_REGION
    GARDENER_KUBECONFIG
    GARDENER_PROJECT_NAME
    GARDENER_PROVIDER_SECRET_NAME
    GARDENER_CLUSTER_VERSION
    MACHINE_TYPE
    DISK_SIZE
    DISK_TYPE
    SCALER_MAX
    SCALER_MIN
)

requiredFiles=(
    GARDENER_KUBECONFIG
)

check_required_vars "${requiredVars[@]}"
check_required_files "${requiredFiles[@]}"

echo "Started cluster provisioning, name: ${CLUSTER_NAME}, preset ${GARDENER_CONFIGURATION}"

if [ ! -f "${preset_dir}/shoot.yaml" ]; then
    >&2 echo "File '${preset_dir}/shoot.yaml' required but not found"
    echo "::endgroup::"
    exit 2
fi

# render and apply shoot template
shoot_template=$(envsubst < "${preset_dir}/shoot.yaml")
echo "Trying to apply shoot template into seed cluster"
retries=0
until (echo "$shoot_template" | kubectl --kubeconfig "${GARDENER_KUBECONFIG}" apply -f -); do
  retries+=1
  if [[ retries -gt 2 ]]; then
    echo "Could not apply shoot spec after 3 tries, exiting"
    echo "::endgroup::"
    exit 3
  fi
  echo "Failed, retrying in 15s"
  sleep 15
done
echo "Shoot template applied"

echo "Waiting for shoot operations to be completed..."
kubectl_wait_code=0
kubectl wait --kubeconfig "${GARDENER_KUBECONFIG}" --for=jsonpath='{.status.lastOperation.state}'=Succeeded --timeout=30m "shoots/${CLUSTER_NAME}" || kubectl_wait_code=$?
if [ "${kubectl_wait_code}" -ne 0 ]; then
  echo "Timed out waiting for the shoot provisioning, kubectl exit code: ${kubectl_wait_code}"
  echo "Shoot last operation:"
  kubectl --kubeconfig "${GARDENER_KUBECONFIG}" get shoot "${CLUSTER_NAME}" -o jsonpath='{.status.lastOperation}' | jq
  echo "Shoot status conditions:"
  kubectl --kubeconfig "${GARDENER_KUBECONFIG}" get shoot "${CLUSTER_NAME}" -o jsonpath='{.status.conditions}' | jq
  echo "::endgroup::"
  exit 4
fi

# create kubeconfig request, that creates a kubeconfig which is valid for one day
echo "Storing kubeconfig in ${CLUSTER_KUBECONFIG}"
kubectl create  --kubeconfig "${GARDENER_KUBECONFIG}" \
    -f <(printf '{"spec":{"expirationSeconds":86400}}') \
    --raw "/apis/core.gardener.cloud/v1beta1/namespaces/garden-${GARDENER_PROJECT_NAME}/shoots/${CLUSTER_NAME}/adminkubeconfig" | \
    jq -r ".status.kubeconfig" | \
    base64 -d > "${CLUSTER_KUBECONFIG}"

# The kyma-provisioning-info ConfigMap gates dualstack in istio-manager. On real
# BTP clusters infrastructure-manager writes it; Gardener-native CI shoots must
# fill it in themselves. It belongs to provisioning, so it is created once here,
# and dualStackIPEnabled follows the shoot's IP stack (dualstack shoots enable it).
echo "Creating kyma-provisioning-info ConfigMap"
[ "${GARDENER_IP_STACK}" = "dualstack" ] && dual_stack_enabled=true || dual_stack_enabled=false
kubectl --kubeconfig "${CLUSTER_KUBECONFIG}" create namespace kyma-system \
    --dry-run=client -o yaml | kubectl --kubeconfig "${CLUSTER_KUBECONFIG}" apply -f -
printf 'networkDetails:\n  dualStackIPEnabled: %s\n' "${dual_stack_enabled}" \
    | kubectl --kubeconfig "${CLUSTER_KUBECONFIG}" create configmap kyma-provisioning-info -n kyma-system \
        --from-file=details=/dev/stdin --dry-run=client -o yaml \
    | kubectl --kubeconfig "${CLUSTER_KUBECONFIG}" apply -f -

echo "Shoot provisioning finished"
echo "::endgroup::"
