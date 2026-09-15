#!/usr/bin/env bash

# Description: This script deletes the Gardener cluster
# It requires the following env variables:
# - CLUSTER_NAME - name of the cluster to be deleted
# - GARDENER_KUBECONFIG - Gardener kubeconfig path
# - GARDENER_PROJECT_NAME - name of the Gardener project

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars CLUSTER_NAME GARDENER_PROJECT_NAME
require_files GARDENER_KUBECONFIG

echo "Deprovisioning cluster: ${CLUSTER_NAME}"

kubectl annotate shoot "${CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true \
    --overwrite \
    -n "garden-${GARDENER_PROJECT_NAME}" \
    --kubeconfig "${GARDENER_KUBECONFIG}"

kubectl delete shoot "${CLUSTER_NAME}" \
  --wait="false" \
  --kubeconfig "${GARDENER_KUBECONFIG}" \
  -n "garden-${GARDENER_PROJECT_NAME}"
