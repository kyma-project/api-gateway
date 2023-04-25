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

if [[ -z ${TEST_OIDC_ISSUER_URL} ]]; then
  >&2 echo "Environment variable TEST_OIDC_ISSUER_URL is required but not set"
  exit 2
fi

if [[ -z ${TEST_CLIENT_ID} ]]; then
  >&2 echo "Environment variable TEST_CLIENT_ID is required but not set"
  exit 2
fi

if [[ -z ${TEST_CLIENT_SECRET} ]]; then
  >&2 echo "Environment variable TEST_CLIENT_SECRET is required but not set"
  exit 2
fi

export TEST_REQUEST_TIMEOUT="180"
export TEST_REQUEST_DELAY="2"
export TEST_CLIENT_TIMEOUT=30s
export TEST_CONCURENCY="1"
export EXPORT_RESULT="true"
