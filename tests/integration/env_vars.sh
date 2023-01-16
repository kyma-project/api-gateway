#!/usr/bin/env bash

if [[ -z ${KYMA_DOMAIN} ]]; then
  echo "KYMA_DOMAIN not exported, fallback to default k3d local.kyma.dev"
  export KYMA_DOMAIN=local.kyma.dev
fi

export TEST_HYDRA_ADDRESS="https://oauth2.${KYMA_DOMAIN}"
export TEST_REQUEST_TIMEOUT="120"
export TEST_REQUEST_DELAY="10"
export TEST_DOMAIN="${KYMA_DOMAIN}"
export TEST_CLIENT_TIMEOUT=30s
export TEST_CONCURENCY="8"
export EXPORT_RESULT="true"
