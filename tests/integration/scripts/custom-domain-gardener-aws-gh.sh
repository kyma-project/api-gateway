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
export GARDENER_REGION="eu-west-1"
export GARDENER_PROVIDER_SECRET_NAME="aws-gardener-access"
export GARDENER_PROJECT_NAME="goats"
export GARDENER_CLUSTER_VERSION="1.27.8"

./tests/integration/scripts/custom-domain-gardener-gh.sh
