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

export MACHINE_TYPE="m5.xlarge"
export DISK_SIZE=50
export DISK_TYPE="gp2"
export SCALER_MAX=3
export SCALER_MIN=1
export GARDENER_PROVIDER="aws"
export GARDENER_REGION="europe-west4"
export GARDENER_ZONES="europe-west4-b"

./tests/integration/scripts/test-gardener.sh
