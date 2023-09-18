#!/usr/bin/env bash

LATEST_RELEASE=$2 # for testability

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

RELEASE_TAG=$1

REPOSITORY=${REPOSITORY:-kyma-project/api-gateway}
GITHUB_URL=https://api.github.com/repos/${REPOSITORY}
GITHUB_AUTH_HEADER="Authorization: token ${GITHUB_TOKEN}"
CHANGELOG_FILE="CHANGELOG.md"

if [ "${LATEST_RELEASE}"  == "" ]
then
  LATEST_RELEASE=$(curl -H "${GITHUB_AUTH_HEADER}" -sS "${GITHUB_URL}/releases/latest" | jq -r '.tag_name')
fi

echo -e "\n**Full changelog:** https://github.com/$REPOSITORY/compare/${LATEST_RELEASE}...${RELEASE_TAG}" >> ${CHANGELOG_FILE}
