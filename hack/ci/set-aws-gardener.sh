# This file contains variables for integration tests run on a Gardener in AWS
# It is automatically loaded by provision-gardener.sh and deprovision-gardener.sh

MACHINE_TYPE="m5.xlarge"
DISK_SIZE=50
DISK_TYPE="gp2"
SCALER_MAX=3
SCALER_MIN=1
GARDENER_PROVIDER="aws"
GARDENER_REGION="eu-west-1"
GARDENER_PROVIDER_SECRET_NAME="aws-gardener-access"
GARDENER_CLUSTER_VERSION="1.31.6"
