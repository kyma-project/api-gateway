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
#
#Please look in each provider script for provider specific requirements

# exit on error, and raise error when variable is not set when used
set -e

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

# nice cleanup on exit, be it successful or on fail
trap cleanup EXIT INT

echo "--> Install kyma CLI locally to /tmp/bin"
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/latest/download/kyma_linux_x86_64.tar.gz" \
&& tar -zxvf kyma.tar.gz && chmod +x kyma \
&& rm -f kyma.tar.gz

chmod +x kyma

# Add pwd to path to be able to use Kyma binary
export PATH="${PATH}:${PWD}"
kyma version --client

CLUSTER_NAME=$(LC_ALL=C tr -dc 'a-z' < /dev/urandom | head -c10)

kyma provision gardener gcp \
        --secret "${GARDENER_KYMA_PROW_PROVIDER_SECRET_NAME}" \
        --name "${CLUSTER_NAME}" \
        --project "${GARDENER_KYMA_PROW_PROJECT_NAME}" \
        --credentials "${GARDENER_KYMA_PROW_KUBECONFIG}" \
        --region "${GARDENER_REGION}" \
        -z "${GARDENER_ZONES}" \
        -t "n1-standard-4" \
        --scaler-max 3 \
        --scaler-min 1 \
        --kube-version="${GARDENER_CLUSTER_VERSION}" \
        --attempts 1 \
        --verbose

export KYMA_DOMAIN="${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"

./tests/integration/scripts/jobguard/run.sh

make install-kyma
make test-integration

