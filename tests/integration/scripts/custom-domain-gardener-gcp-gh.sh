#!/usr/bin/env bash

#
##Description: This scripts installs and tests api-gateway custom domain test as well as gateway test using the CLI on a real Gardener GCP cluster.
## exit on error, and raise error when variable is not set when used

set -e

export MACHINE_TYPE="n1-standard-4"
export DISK_SIZE=50
export DISK_TYPE="pd-standard"
export SCALER_MAX=3
export SCALER_MIN=1
export GARDENER_PROVIDER="gcp"
export GARDENER_REGION="europe-west4"
export GARDENER_ZONES="europe-west4-c,europe-west4-b,europe-west4-a"
export GARDENER_PROVIDER_SECRET_NAME="trial-secretbinding-gcp"
export GARDENER_PROJECT_NAME="goatz"
export GARDENER_CLUSTER_VERSION="1.26.9"

./tests/integration/scripts/custom-domain-gardener-gh.sh
