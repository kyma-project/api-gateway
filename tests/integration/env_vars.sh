#!/bin/bash

set -eo pipefail

if [[ -z ${KYMA_DOMAIN} ]]; then
  >&2 echo "Environment variable KYMA_DOMAIN not set, fallback to default k3d 'local.kyma.dev'"
  export KYMA_DOMAIN="local.kyma.dev"
fi

export TEST_OIDC_CONFIG_URL="${OIDC_CONFIG_URL}"
export TEST_CLIENT_ID="${CLIENT_ID}"
export TEST_CLIENT_SECRET="${CLIENT_SECRET}"
export TEST_REQUEST_DELAY="10"
export TEST_DOMAIN="${KYMA_DOMAIN}"
export EXPORT_RESULT="true"
export IS_GARDENER="${IS_GARDENER}"
