#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

# Ensure RELEASE_TAG is provided
RELEASE_TAG=$1
if [ -z "$RELEASE_TAG" ]; then
  echo "Error: RELEASE_TAG is not set."
  exit 1
fi

# Set default repository if not provided
REPOSITORY=${REPOSITORY:-kyma-project/api-gateway}
GITHUB_URL=https://api.github.com/repos/${REPOSITORY}
GITHUB_AUTH_HEADER="Authorization: token ${GITHUB_TOKEN}"
CHANGELOG_FILE="CHANGELOG.md"

# Extract major, minor, and patch version numbers
set -- ${RELEASE_TAG//./ }
MAJOR=$1
MINOR=$2
PATCH=$3

# Fetch all release tags
TAGS=$(curl -s -H "$GITHUB_AUTH_HEADER" "$GITHUB_URL/releases" | grep -o '"tag_name": "[^"]*"' | sed -E 's/"tag_name": "([^"]*)"/\1/' | sort -V)

# Determine the previous release based on versioning rules
LATEST_TAG=""
if [ $PATCH -ne 0 ]; then
  LATEST_TAG=$(echo "$TAGS" | grep -E "^${MAJOR}\.${MINOR}\." | tail -1)
elif [ "$MINOR" -ne 0 ]; then
  PREV_MINOR=$(($MINOR - 1))
  LATEST_TAG=$(echo "$TAGS" | grep -E "^${MAJOR}\.${PREV_MINOR}\." | head -n 1)
else
  PREV_MAJOR=$(($MAJOR - 1))
  LATEST_TAG=$(echo "$TAGS" | grep -E "^${PREV_MAJOR}\." | head -n 1)
fi

# Fetch commit history between the previous and current release
echo -e "\n**Full changelog:** https://github.com/$REPOSITORY/compare/${LATEST_TAG}...${RELEASE_TAG}" >> ${CHANGELOG_FILE}