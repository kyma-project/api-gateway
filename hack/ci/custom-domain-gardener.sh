#!/usr/bin/env bash

#
##Description: This scripts installs and tests api-gateway custom domain test as well as gateway test using the CLI on a real Gardener GCP cluster.
## exit on error, and raise error when variable is not set when used
## IMG env variable expected (for make deploy), which points to the image in the registry

set -eo pipefail

if [ $# -lt 1 ]; then
    >&2 echo "Make target is required as parameter"
    exit 2
fi

function check_required_vars() {
  local requiredVarMissing=false
  for var in "$@"; do
    if [ -z "${!var}" ]; then
      >&2 echo "Environment variable ${var} is required but not set"
      requiredVarMissing=true
    fi
  done
  if [ "${requiredVarMissing}" = true ] ; then
    exit 2
  fi
}

function check_required_files() {
  local requiredFileMissing=false
  for file in "$@"; do
    path=$(eval echo "\$$file")
    if [ ! -f "${path}" ]; then
        >&2 echo "File '${path}' required but not found"
        requiredFileMissing=true
    fi
  done
  if [ "${requiredFileMissing}" = true ] ; then
    exit 2
  fi
}

requiredVars=(
    GARDENER_KUBECONFIG
    GARDENER_PROJECT_NAME
    CLIENT_ID
    CLIENT_SECRET
    OIDC_CONFIG_URL
    TEST_SA_ACCESS_KEY_PATH
)

requiredFiles=(
    GARDENER_KUBECONFIG
    TEST_SA_ACCESS_KEY_PATH
)

check_required_vars "${requiredVars[@]}"
check_required_files "${requiredFiles[@]}"

function cleanup() {
  kubectl annotate shoot "${CLUSTER_NAME}" confirmation.gardener.cloud/deletion=true \
      --overwrite \
      -n "garden-${GARDENER_PROJECT_NAME}" \
      --kubeconfig "${GARDENER_KUBECONFIG}"

  kubectl delete shoot "${CLUSTER_NAME}" \
    --wait="false" \
    --kubeconfig "${GARDENER_KUBECONFIG}" \
    -n "garden-${GARDENER_PROJECT_NAME}"
}

# Cleanup on exit, be it successful or on fail
trap cleanup EXIT INT

# Add pwd to path to be able to use binaries downloaded in scripts
export PATH="${PATH}:${PWD}"

CLUSTER_NAME=ag-$(echo $RANDOM | md5sum | head -c 7)
export CLUSTER_NAME

TMP_FOLDER=$(mktemp -d)
export CLUSTER_KUBECONFIG="${TMP_FOLDER}/${CLUSTER_NAME}_kubeconfig.yaml"

./hack/ci/provision-gardener.sh

export KUBECONFIG="${CLUSTER_KUBECONFIG}"

echo "installing istio"
make install-istio

echo "deploying api-gateway"
make deploy

# KYMA_DOMAIN is required by the tests
export TEST_DOMAIN="${CLUSTER_NAME}.${GARDENER_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"
export KYMA_DOMAIN="${TEST_DOMAIN}"
export TEST_CUSTOM_DOMAIN="goat.build.kyma-project.io"
export IS_GARDENER=true

echo "waiting for the ingress gateway external address"
[ "$GARDENER_PROVIDER" == "aws" ] && address_field="{.status.loadBalancer.ingress[0].hostname}" || address_field="{.status.loadBalancer.ingress[0].ip}"
kubectl wait --timeout=300s --namespace istio-system services/istio-ingressgateway --for=jsonpath="${address_field}"
ingress_external_address=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath="${address_field}")
ingress_external_status_port=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.spec.ports[?(@.name=="status-port")].targetPort}')

echo "determined ingress external address: ${ingress_external_address} and external status port: ${ingress_external_status_port}"

echo "waiting until it is possible to connect to the ingress gateway"
trial=1
# check if it is possible to establish connection to the ingress gateway (the exact http status code doesn't matter)
until curl --silent --output /dev/null "http://${ingress_external_address}:${ingress_external_status_port}"
do
  if (( trial >= 60 ))
  then
     echo "exceeded number of trials while waiting for the ingress gateway, giving up..."
     exit 4
  fi
  echo "ingress gateway does not respond, trying again..."
  sleep 10
  trial=$((trial + 1))
done

for make_target in "$@"
do
    make $make_target
done
