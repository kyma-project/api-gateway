#!/usr/bin/env bash

# Description: This script provisions a Gardener cluster
# It requires the following env variables:
# - CLUSTER_NAME - name of the cluster to be created
# - CLUSTER_KUBECONFIG - target path where the kubeconfig of the newly created cluster is stored
# - GARDENER_KUBECONFIG - Gardener kubeconfig path
# - GARDENER_PROVIDER - provider name (cloud name) used to create cluster
# - GARDENER_PROJECT_NAME - name of the Gardener project
# Other variables are loaded from set-${GARDENER_PROVIDER}-gardener.sh script

set -eo pipefail
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
    exit 2
  fi
}

check_required_vars GARDENER_PROVIDER
if [ ! -f "${script_dir}/set-${GARDENER_PROVIDER}-gardener.sh" ]; then
    >&2 echo "File '${script_dir}/set-${GARDENER_PROVIDER}-gardener.sh' required but not found"
    exit 2
fi
set -a # autoexport variables in the sourced file
source "${script_dir}/set-${GARDENER_PROVIDER}-gardener.sh"
set +a

requiredVars=(
    CLUSTER_NAME
    CLUSTER_KUBECONFIG
    GARDENER_PROVIDER
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

echo "Started cluster provisioning, name: ${CLUSTER_NAME}, provider ${GARDENER_PROVIDER}"

if [ ! -f "${script_dir}/shoot_${GARDENER_PROVIDER}.yaml" ]; then
    >&2 echo "File '${script_dir}/shoot_${GARDENER_PROVIDER}.yaml' required but not found"
    exit 2
fi

# render and apply shoot template
shoot_template=$(envsubst < "${script_dir}/shoot_${GARDENER_PROVIDER}.yaml")
echo "Trying to apply shoot template into seed cluster"
retries=0
until (echo "$shoot_template" | kubectl --kubeconfig "${GARDENER_KUBECONFIG}" apply -f -); do
  retries+=1
  if [[ retries -gt 2 ]]; then
    echo "Could not apply shoot spec after 3 tries, exiting"
    exit 1
  fi
  echo "Failed, retrying in 15s"
  sleep 15
done
echo "Shoot template applied"

echo "Waiting for cluster to be ready..."
kubectl wait  --kubeconfig "${GARDENER_KUBECONFIG}" --for=condition=EveryNodeReady shoot/${CLUSTER_NAME} --timeout=25m
# create kubeconfig request, that creates a kubeconfig which is valid for one day

echo "Storing kubeconfig in ${CLUSTER_KUBECONFIG}"
kubectl create  --kubeconfig "${GARDENER_KUBECONFIG}" \
    -f <(printf '{"spec":{"expirationSeconds":86400}}') \
    --raw "/apis/core.gardener.cloud/v1beta1/namespaces/garden-${GARDENER_PROJECT_NAME}/shoots/${CLUSTER_NAME}/adminkubeconfig" | \
    jq -r ".status.kubeconfig" | \
    base64 -d > "${CLUSTER_KUBECONFIG}"

echo "Wait until API Server is ready"
# wait until apiserver /readyz endpoint returns "ok"
timeout=0
until (kubectl --kubeconfig "${CLUSTER_KUBECONFIG}" get --raw "/readyz"); do
  timeout+=1
  # 10 minutes
  if [[ $timeout -gt 600 ]]; then
    echo "Timed out waiting for API Server to be ready"
    exit 1
  fi
  sleep 1
done
echo "API Server is ready"

echo "Waiting for shoot operations to be completed..."
kubectl wait --kubeconfig "${GARDENER_KUBECONFIG}" --for=jsonpath='{.status.lastOperation.state}'=Succeeded --timeout=600s "shoots/${CLUSTER_NAME}"
echo "Shoot provisioning finished"
