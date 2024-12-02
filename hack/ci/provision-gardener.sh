#!/usr/bin/env bash

#
##Description: This script provisions a Gardener cluster with config specified in environmental variables

set -eo pipefail

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

if [ ! -f "./hack/ci/shoot_${GARDENER_PROVIDER}.yaml" ]; then
    >&2 echo "File './hack/ci/shoot_${GARDENER_PROVIDER}.yaml' required but not found"
    exit 2
fi

# render and applyshoot template
shoot_template=$(envsubst < "./hack/ci/shoot_${GARDENER_PROVIDER}.yaml")

echo "trying to apply shoot template into seed cluster"
retries=0
until (echo "$shoot_template" | kubectl --kubeconfig "${GARDENER_KUBECONFIG}" apply -f -); do
  retries+=1
  if [[ retries -gt 2 ]]; then
    echo "could not apply shoot spec after 3 tries, exiting"
    exit 1
  fi
  echo "failed, retrying in 15s"
  sleep 15
done

echo "waiting for cluster to be ready..."
kubectl wait  --kubeconfig "${GARDENER_KUBECONFIG}" --for=condition=EveryNodeReady shoot/${CLUSTER_NAME} --timeout=25m
# create kubeconfig request, that creates a kubeconfig which is valid for one day
kubectl create  --kubeconfig "${GARDENER_KUBECONFIG}" \
    -f <(printf '{"spec":{"expirationSeconds":86400}}') \
    --raw "/apis/core.gardener.cloud/v1beta1/namespaces/garden-${GARDENER_PROJECT_NAME}/shoots/${CLUSTER_NAME}/adminkubeconfig" | \
    jq -r ".status.kubeconfig" | \
    base64 -d > "${CLUSTER_KUBECONFIG}"

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

echo "waiting for shoot operations to be completed..."
kubectl wait --kubeconfig "${GARDENER_KUBECONFIG}" --for=jsonpath='{.status.lastOperation.state}'=Succeeded --timeout=600s "shoots/${CLUSTER_NAME}"
