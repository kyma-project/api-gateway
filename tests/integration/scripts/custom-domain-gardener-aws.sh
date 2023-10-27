#!/usr/bin/env bash

#
##Description: This scripts installs and tests api-gateway custom domain test as well as gateway test using the CLI on a real Gardener AWS cluster.
## exit on error, and raise error when variable is not set when used

set -e

export MACHINE_TYPE="m5.xlarge"
export DISK_SIZE=50
export DISK_TYPE="gp2"
export SCALER_MAX=3
export SCALER_MIN=1
export GARDENER_PROVIDER="aws"
export GARDENER_REGION="eu-north-1"
export GARDENER_ZONES="eu-north-1b,eu-north-1c,eu-north-1a"

./tests/integration/scripts/custom-domain-gardener.sh
