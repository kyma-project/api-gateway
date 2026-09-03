# Variables for e2e tests on a Gardener AWS dualstack shoot.
# Auto-loaded by provision.sh when GARDENER_CONFIGURATION=aws-dualstack. Shoot template
# lives in shoot.yaml and enables VPC dual-stack + AWS LBC.

MACHINE_TYPE="m5.xlarge"
DISK_SIZE=50
DISK_TYPE="gp2"
SCALER_MAX=3
SCALER_MIN=3 # we need at least 3 machines, because there are 3 availability zones
GARDENER_PROVIDER="aws"
GARDENER_IP_STACK="dualstack"
GARDENER_REGION="eu-west-1"
GARDENER_PROVIDER_SECRET_NAME="goat-aws-secret"
GARDENER_CLUSTER_VERSION="1.33.3"
