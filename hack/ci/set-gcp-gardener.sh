# This file contains variables for integration tests run on a Gardener in GCP
# It is automatically loaded by provision-gardener.sh and deprovision-gardener.sh

MACHINE_TYPE="n2-standard-4"
DISK_SIZE=50
DISK_TYPE="pd-standard"
SCALER_MAX=3
SCALER_MIN=1
GARDENER_PROVIDER="gcp"
GARDENER_REGION="europe-west3"
GARDENER_PROVIDER_SECRET_NAME="goat"
GARDENER_CLUSTER_VERSION="1.30.8"
