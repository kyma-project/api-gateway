#!/usr/bin/env bash

#
##Description: This script provisions a Gardener cluster with config specified in environmental variables

set -euo pipefail

function check_required_vars() {
  local requiredVarMissing=false
  for var in "$@"; do
    if [ -z "${var}" ]; then
      >&2 echo "Environment variable ${var} is required but not set"
      requiredVarMissing=true
    fi
  done
  if [ "${requiredVarMissing}" = true ] ; then
    exit 2
  fi
}

requiredVars=(
    CLUSTER_NAME
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

check_required_vars "${requiredVars[@]}"

# render and applyshoot template
shoot_template=$(envsubst < ./tests/integration/scripts/shoot_${GARDENER_PROVIDER}.yaml)

echo "$shoot_template" | kubectl --kubeconfig "${GARDENER_KUBECONFIG}" apply -f -

echo "waiting fo cluster to be ready..."
kubectl wait  --kubeconfig "${GARDENER_KUBECONFIG}" --for=condition=EveryNodeReady shoot/${CLUSTER_NAME} --timeout=17m
# create kubeconfig request, that creates a kubeconfig which is valid for one day
kubectl create  --kubeconfig "${GARDENER_KUBECONFIG}" \
    -f <(printf '{"spec":{"expirationSeconds":86400}}') \
    --raw "/apis/core.gardener.cloud/v1beta1/namespaces/garden-${GARDENER_PROJECT_NAME}/shoots/${CLUSTER_NAME}/adminkubeconfig" | \
    jq -r ".status.kubeconfig" | \
    base64 -d > "${CLUSTER_NAME}_kubeconfig.yaml"

# wait until apiserver /readyz endpoint returns "ok"
isOK=""
timeout=0
until [[ $isOK == "ok" ]]; do
  isOK=$(kubectl --kubeconfig "${CLUSTER_NAME}_kubeconfig.yaml" get --raw "/readyz")
  # 5 minutes
  if [[ $timeout -gt 300 ]]; then
    echo "Timed out waiting for API Server to be ready"
    exit 1
  fi
  timeout+=1
  sleep 1
done

# replace the default kubeconfig
mkdir -p ~/.kube
mv "${CLUSTER_NAME}_kubeconfig.yaml" ~/.kube/config

