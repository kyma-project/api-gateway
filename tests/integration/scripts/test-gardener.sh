#!/usr/bin/env bash

#Description: This scripts installs and test Kyma using the CLI on a real Gardener AWS cluster.
#
#Expected common vars:
# - GARDENER_REGION - Gardener compute region
# - GARDENER_ZONES - Gardener compute zones inside the region
# - GARDENER_CLUSTER_VERSION - Version of the Kubernetes cluster
# - GARDENER_KYMA_PROW_KUBECONFIG - Kubeconfig of the Gardener service account
# - GARDENER_KYMA_PROW_PROJECT_NAME - Name of the gardener project where the cluster will be integrated.
# - GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME - Name of the secret configured in the gardener project to access the cloud provider
#

# exit on error, and raise error when variable is not set when used
set -e

CLUSTER_NAME=$(LC_ALL=C tr -dc 'a-z' < /dev/urandom | head -c10)
export CLUSTER_NAME

cleanup() {
kubectl annotate shoot "${CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true \
    --overwrite \
    -n "garden-${GARDENER_KYMA_PROW_PROJECT_NAME}" \
    --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}"

kubectl delete shoot "${CLUSTER_NAME}" \
  --wait="false" \
  --kubeconfig "${GARDENER_KYMA_PROW_KUBECONFIG}" \
  -n "garden-${GARDENER_KYMA_PROW_PROJECT_NAME}"
}

# Cleanup on exit, be it successful or on fail
trap cleanup EXIT INT

./tests/integration/scripts/provision-gardener.sh

./tests/integration/scripts/jobguard.sh
make install-kyma

# KYMA_DOMAIN is required by the tests
export KYMA_DOMAIN="${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"
make test-integration

