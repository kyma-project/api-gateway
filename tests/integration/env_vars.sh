#!/bin/bash

set -eo pipefail

if [[ -z ${KYMA_DOMAIN} ]]; then
  >&2 echo "Environment variable KYMA_DOMAIN not set, fallback to default k3d 'local.kyma.dev'"
  export KYMA_DOMAIN="local.kyma.dev"
fi

if [[ -z ${IAS_ADDRESS} ]]; then
  >&2 echo "Environment variable IAS_ADDRESS is required but not set"
  exit 2
fi

if [[ -z ${CLIENT_ID} || -z ${CLIENT_SECRET} ]]; then
  >&2 echo "Environment variable CLIENT_ID and CLIENT_SECRET are required but not set"
  exit 2
fi

export TEST_IAS_ADDRESS="${IAS_ADDRESS}"
export TEST_CLIENT_ID="${CLIENT_ID}"
export TEST_CLIENT_SECRET="${CLIENT_SECRET}"
export TEST_REQUEST_TIMEOUT="120"
export TEST_REQUEST_DELAY="10"
export TEST_DOMAIN="${KYMA_DOMAIN}"
export TEST_CLIENT_TIMEOUT=30s
export TEST_CONCURENCY="8"
export EXPORT_RESULT="true"
