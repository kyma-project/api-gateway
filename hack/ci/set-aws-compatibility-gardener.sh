# This file contains variables for integration tests run on a Gardener in AWS
# It is automatically loaded by provision-gardener.sh and deprovision-gardener.sh
# This particular one is used for running compatibility tests with future version of Kubernetes

MACHINE_TYPE="m5.xlarge"
DISK_SIZE=50
DISK_TYPE="gp2"
SCALER_MAX=3
SCALER_MIN=3
GARDENER_PROVIDER="aws"
GARDENER_REGION="eu-west-1"
GARDENER_PROVIDER_SECRET_NAME="aws-gardener-access"
# Should use the latest available Kubernetes version for AWS on Gardener
GARDENER_CLUSTER_VERSION="1.33.4"
