# Variables for e2e tests on a Gardener AWS dualstack shoot.
# Auto-loaded by provision-gardener.sh / deprovision-gardener.sh when
# GARDENER_PROVIDER=aws-dualstack. Shoot template lives in
# shoot_aws-dualstack.yaml and enables VPC dual-stack + AWS LBC.

MACHINE_TYPE="m5.xlarge"
DISK_SIZE=50
DISK_TYPE="gp2"
SCALER_MAX=3
SCALER_MIN=3 # 3 machines to spread across 3 zones with HA control plane
GARDENER_PROVIDER="aws-dualstack"
GARDENER_REGION="eu-west-1"
GARDENER_PROVIDER_SECRET_NAME="goat-aws-secret"
GARDENER_CLUSTER_VERSION="1.35.5"
