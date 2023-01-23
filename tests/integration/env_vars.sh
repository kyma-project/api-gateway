#!/bin/bash

set -eo pipefail

if [[ -z ${KYMA_DOMAIN} ]]; then
  >&2 echo "Environment variable KYMA_DOMAIN not set, fallback to default k3d 'local.kyma.dev'"
  echo export KYMA_DOMAIN="local.kyma.dev"
fi

if [[ -z ${CLIENT_ID} || -z ${CLIENT_SECRET} ]]; then
  >&2 echo "Environment variable CLIENT_ID and CLIENT_SECRET are required but not set"
  exit 2
fi

echo export TEST_IAS_ADDRESS="https://kymagoattest.accounts400.ondemand.com"
echo export TEST_CLIENT_ID="${CLIENT_ID}"
echo export TEST_CLIENT_SECRET="${CLIENT_SECRET}"
echo export TEST_REQUEST_TIMEOUT="120"
echo export TEST_REQUEST_DELAY="10"
echo export TEST_DOMAIN="${KYMA_DOMAIN}"
echo export TEST_CLIENT_TIMEOUT=30s
echo export TEST_CONCURENCY="8"
echo export EXPORT_RESULT="true"
