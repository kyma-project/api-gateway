#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

TARGET_BRANCH="$1"

# If the first tag in branch history doesn't match release tag minor (or target is main)
# install api-gateway from the latest release instead
TAG=$(git describe --tags --abbrev=0)
if [ "${TAG%.*}" == "${TARGET_BRANCH#release\-}" ]
then
  echo "Installing APIGateway ${TAG}"
  RELEASE_MANIFEST_URL="https://github.com/kyma-project/api-gateway/releases/download/${TAG}/api-gateway-manager.yaml"
  curl -L "$RELEASE_MANIFEST_URL" | kubectl apply -f -
else
  echo "Installing APIGateway from latest release"
  RELEASE_MANIFEST_URL="https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml"
  curl -L "$RELEASE_MANIFEST_URL" | kubectl apply -f -
fi