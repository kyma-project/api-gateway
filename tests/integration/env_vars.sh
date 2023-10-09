#!/bin/bash

set -eo pipefail

if [[ -z ${KYMA_DOMAIN} ]]; then
  >&2 echo "Environment variable KYMA_DOMAIN not set, fallback to default k3d 'local.kyma.dev'"
  export KYMA_DOMAIN="local.kyma.dev"
fi

if [[ -z ${OIDC_ISSUER_URL} ]]; then
  >&2 echo "Environment variable OIDC_ISSUER_URL is required but not set"
  exit 2
fi

if [[ -z ${CLIENT_ID} || -z ${CLIENT_SECRET} ]]; then
  >&2 echo "Environment variable CLIENT_ID and CLIENT_SECRET are required but not set"
  exit 2
fi

export TEST_OIDC_ISSUER_URL="${OIDC_ISSUER_URL}"
export TEST_CLIENT_ID="${CLIENT_ID}"
export TEST_CLIENT_SECRET="${CLIENT_SECRET}"
export TEST_REQUEST_DELAY="10"
export TEST_DOMAIN="${KYMA_DOMAIN}"
export EXPORT_RESULT="true"
export IS_GARDENER="${IS_GARDENER}"
