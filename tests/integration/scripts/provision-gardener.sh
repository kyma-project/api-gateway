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
    GARDENER_ZONES
    GARDENER_KYMA_PROW_KUBECONFIG
    GARDENER_KYMA_PROW_PROJECT_NAME
    GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME
    GARDENER_CLUSTER_VERSION
    MACHINE_TYPE
    DISK_SIZE
    DISK_TYPE
    SCALER_MAX
    SCALER_MIN
)

check_required_vars "${requiredVars[@]}"

# Install Kyma CLI in latest version
echo "--> Install kyma CLI locally to /tmp/bin"
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/latest/download/kyma_linux_x86_64.tar.gz" \
&& tar -zxvf kyma.tar.gz && chmod +x kyma \
&& rm -f kyma.tar.gz

kyma version --client
kyma provision gardener ${GARDENER_PROVIDER} \
        --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" \
        --name "${CLUSTER_NAME}" \
        --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
        --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" \
        --zones "${GARDENER_ZONES}" \
        --type "${MACHINE_TYPE}" \
        --disk-size $DISK_SIZE \
        --disk-type "${DISK_TYPE}" \
        --scaler-max $SCALER_MAX \
        --scaler-min $SCALER_MIN \
        --kube-version="${GARDENER_CLUSTER_VERSION}" \
        --attempts 3 \
        --verbose