#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail


# Uncomment and eventually adjust commented code bellow after first API Gateway Manager release
#TARGET_BRANCH="$1"

# The if statement is here to make the script work properly on release branches
# Example: if we want to patch 0.2.0 to 0.2.1 when already 0.3.0 exists it makes
# the script deploy 0.2.0 instead 0.3.0 so we don't test upgrade from 0.3.0 to 0.2.1
#if [ "$TARGET_BRANCH" != "main" ] && [ "$TARGET_BRANCH" != "" ]
#then
#  TAG=$(git describe --tags --abbrev=0)
#  RELEASE_MANIFEST_URL="https://github.com/kyma-project/api-gateway/releases/download/${TAG}/api-gateway-manager.yaml"
#  curl -L "$RELEASE_MANIFEST_URL" | kubectl apply -f -
#else
#  RELEASE_MANIFEST_URL="https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml"
#  curl -L "$RELEASE_MANIFEST_URL" | kubectl apply -f -
#fi

git checkout mod-dev
MOD_DEV_SHA=$(git log -n 1 --pretty=format:"%H" mod-dev)
IMG=api-gateway-controller:v$(git show -s --date=format:'%Y%m%d' --format=%cd "$MOD_DEV_SHA")-$(printf %.8s "$MOD_DEV_SHA")
echo "$IMG"
#make deploy