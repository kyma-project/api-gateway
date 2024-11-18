#!/bin/bash

set -eo pipefail

if [[ -z ${TEST_DOMAIN} ]]; then
  >&2 echo "Environment variable TEST_DOMAIN is required but not set"
  exit 2
fi

if [[ -z ${TEST_CUSTOM_DOMAIN} ]]; then
  >&2 echo "Environment variable TEST_CUSTOM_DOMAIN is required but not set"
  exit 2
fi

if [[ -z ${TEST_SA_ACCESS_KEY_PATH} ]]; then
  >&2 echo "Environment variable TEST_SA_ACCESS_KEY_PATH is required but not set"
  exit 2
fi

if [[ -z ${OIDC_CONFIG_URL} ]]; then
  >&2 echo "Environment variable OIDC_CONFIG_URL is required but not set"
  exit 2
fi

if [[ -z ${CLIENT_ID} ]]; then
  >&2 echo "Environment variable CLIENT_ID is required but not set"
  exit 2
fi

if [[ -z ${CLIENT_SECRET} ]]; then
  >&2 echo "Environment variable CLIENT_SECRET is required but not set"
  exit 2
fi

export TEST_OIDC_CONFIG_URL="${OIDC_CONFIG_URL}"
export TEST_CLIENT_ID="${CLIENT_ID}"
export TEST_CLIENT_SECRET="${CLIENT_SECRET}"
export TEST_CONCURRENCY="1"
export EXPORT_RESULT="true"
export TEST_REQUEST_TIMEOUT="400"
