#!/usr/bin/env bash

#
##Description: This scripts installs and test api-gateway custom domain test using the CLI on a real Gardener GCP cluster.
## exit on error, and raise error when variable is not set when used

set -e

export MACHINE_TYPE="m5.xlarge"
export DISK_SIZE=50
export DISK_TYPE="gp2"
export SCALER_MAX=3
export SCALER_MIN=1
export GARDENER_PROVIDER="aws"
export GARDENER_REGION="eu-central-1"
export GARDENER_ZONES="eu-central-1b,eu-central-1c,eu-central-1a"

./tests/integration/scripts/custom-domain-gardner.sh
