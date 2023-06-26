#!/usr/bin/env bash

#
##Description: This scripts installs and test api-gateway custom domain test using the CLI on a real Gardener GCP cluster.
## exit on error, and raise error when variable is not set when used

set -e

export MACHINE_TYPE="n2-standard-4"
export DISK_SIZE=50
export DISK_TYPE="pd-standard"
export SCALER_MAX=3
export SCALER_MIN=1
export GARDENER_PROVIDER="gcp"
export GARDENER_REGION="europe-west3"
export GARDENER_ZONES="europe-west3-c,europe-west3-b,europe-west3-a"

./tests/integration/scripts/custom-domain-gardner.sh
