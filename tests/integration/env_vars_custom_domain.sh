#!/bin/bash

set -eo pipefail

if [[ -z ${KUBECONFIG} ]]; then
  >&2 echo "Environment variable KUBECONFIG is required but not set"
  exit 2
fi

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

if [[ -n ${OIDC_CONFIG_URL} ]]; then
  if [[ -z ${CLIENT_ID} ]]; then
    >&2 echo "Environment variable CLIENT_ID is required when OIDC_CONFIG_URL is set"
    exit 2
  fi

  if [[ -z ${CLIENT_SECRET} ]]; then
    >&2 echo "Environment variable CLIENT_SECRET is required when OIDC_CONFIG_URL is set"
    exit 2
  fi
fi

export TEST_OIDC_CONFIG_URL="${OIDC_CONFIG_URL}"
export TEST_CLIENT_ID="${CLIENT_ID}"
export TEST_CLIENT_SECRET="${CLIENT_SECRET}"
export TEST_CONCURRENCY="1"
export EXPORT_RESULT="true"
export TEST_REQUEST_ATTEMPTS="120"
