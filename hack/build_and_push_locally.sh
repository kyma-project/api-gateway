#!/usr/bin/env bash

# The script uses the following environment variables (if not provided script defines default values):
#                       LOCAL_REGISTRY
#                       K3D_REGISTRY
#                       PR_NAME
#                       MODULE_PREFIX
# and returns local reference to the created module to stdout
# ./build_and_push_locally.sh

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being maskedPORT=5001

# registry access from local machine
LOCAL_REGISTRY=${LOCAL_REGISTRY:-localhost:5001}

#registry access from k3d cluster
K3D_REGISTRY=${K3D_REGISTRY:-k3d-kyma-registry:5000}

PR_NAME=${PR_NAME:-PR-undefined}
IMG_NAME=api-gateway-manager:${PR_NAME}

MODULE_PREFIX=${MODULE_PREFIX:-0.0.0}
MODULE_VERSION=${MODULE_PREFIX}-${PR_NAME}
EXTENDED_MODULE_VERSION=v${MODULE_VERSION}
MODULE_NAME=component-descriptors/kyma.project.io/module/api-gateway

echo "Creating binary image and pushing to registry: ${LOCAL_REGISTRY}"
make module-image LOCAL_REGISTRY=${LOCAL_REGISTRY} IMG=${LOCAL_REGISTRY}/${IMG_NAME}

echo "Creating OCI module image and pushing to registry (no security scanning configuration in module template): ${LOCAL_REGISTRY}"
make module-build IMG=${K3D_REGISTRY}/${IMG_NAME} MODULE_REGISTRY=${LOCAL_REGISTRY} MODULE_VERSION=${MODULE_VERSION}

MODULE_REFERENCE=${LOCAL_REGISTRY}/${MODULE_NAME}:${EXTENDED_MODULE_VERSION}

echo "Module reference: ${MODULE_REFERENCE}"


