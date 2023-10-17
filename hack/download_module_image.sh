#!/usr/bin/env bash

# This script has the following arguments:
#     the mandatory link to a module image,
#     optional "ci" to indicate call from CI pipeline
# Example:
# ./download_module_image.sh europe-docker.pkg.dev/kyma-project/prod/unsigned/component-descriptors/kyma.project.io/module/api-gateway:v0.2.3 ci

CI=${2-manual}  # if called from any workflow "ci" is expected here

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

IMAGE_NAME=$1
TARGET_DIRECTORY=${TARGET_DIRECTORY:-downloaded_module}

# for local runs
rm -rf ${TARGET_DIRECTORY}

echo -e "\n--- Downloading module image: ${IMAGE_NAME}"
mkdir ${TARGET_DIRECTORY}

# tls setting to allow local access over http, when invoked from CI https is used
TLS_OPTIONS=
if [ "${CI}" != "ci" ]
then
  TLS_OPTIONS=--src-tls-verify=false
fi

skopeo copy ${TLS_OPTIONS} docker://${IMAGE_NAME} dir:${TARGET_DIRECTORY}
