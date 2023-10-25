#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

TARGET_BRANCH="$1"

# The if statement is here to make the script work properly on release branches
# Example: if we want to patch 0.2.0 to 0.2.1 when already 0.3.0 exists it makes
# the script deploy 0.2.0 instead 0.3.0 so we don't test upgrade from 0.3.0 to 0.2.1
if [ "$TARGET_BRANCH" == "release-2.0" ] # Workaround for first API-Gateway module release
then
  RELEASE_MANIFEST_URL="https://github.com/kyma-project/api-gateway/releases/download/2.0.0-rc/api-gateway-manager.yaml"
  echo "Applying release manifest: $RELEASE_MANIFEST_URL"
  curl -L "$RELEASE_MANIFEST_URL" | kubectl apply -f -
elif [ "$TARGET_BRANCH" != "main" ] && [ "$TARGET_BRANCH" != "" ]
then
  TAG=$(git describe --tags --abbrev=0)
  RELEASE_MANIFEST_URL="https://github.com/kyma-project/api-gateway/releases/download/${TAG}/api-gateway-manager.yaml"
  echo "Applying release manifest: $RELEASE_MANIFEST_URL"
  curl -L "$RELEASE_MANIFEST_URL" | kubectl apply -f -
else
  RELEASE_MANIFEST_URL="https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml"
  echo "Applying release manifest: $RELEASE_MANIFEST_URL"
  curl -L "$RELEASE_MANIFEST_URL" | kubectl apply -f -
fi
