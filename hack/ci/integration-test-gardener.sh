#!/usr/bin/env bash

# Description: This script runs given integration tests on a real Gardener cluster
# It installs istio and api gateway and then runs make test targets provided via commandline arguments to that script
# It requires the following env variables:
# - IMG - API gateway image to be deployed (by make deploy)
# - CLUSTER_NAME - Gardener cluster name
# - CLUSTER_KUBECONFIG - Gardener cluster kubeconfig path

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

requiredVars=(
    IMG
    CLUSTER_NAME
    CLUSTER_KUBECONFIG
)

check_required_vars "${requiredVars[@]}"

make_target="$1"

if [ -z "$make_target" ]; then
    echo "Make target is required as parameter"
    exit 3
fi

echo "Make target: $make_target"

echo "Executing tests in cluster ${CLUSTER_NAME}, kubeconfig ${CLUSTER_KUBECONFIG}"
export KUBECONFIG="${CLUSTER_KUBECONFIG}"

export CLUSTER_DOMAIN=$(kubectl get configmap -n kube-system shoot-info -o jsonpath="{.data.domain}")
echo "Cluster domain: ${CLUSTER_DOMAIN}"

export GARDENER_PROVIDER=$(kubectl get configmap -n kube-system shoot-info -o jsonpath="{.data.provider}")
echo "Gardener provider: ${GARDENER_PROVIDER}"

export TEST_DOMAIN="${CLUSTER_DOMAIN}"
export IS_GARDENER=true # this variable is used in tests to make decisions based on the fact that the tests are running in Gardener

# Add pwd to path to be able to use binaries downloaded in scripts
export PATH="${PATH}:${PWD}"

echo "::group::Installing istio"
make install-istio

echo "Deploying api-gateway, image: ${IMG}"
make deploy

echo "Waiting for the ingress gateway external address"
[ "$GARDENER_PROVIDER" == "aws" ] && address_field="{.status.loadBalancer.ingress[0].hostname}" || address_field="{.status.loadBalancer.ingress[0].ip}"
kubectl wait --timeout=300s --namespace istio-system services/istio-ingressgateway --for=jsonpath="${address_field}"
ingress_external_address=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath="${address_field}")
ingress_external_status_port=$(kubectl get services --namespace istio-system istio-ingressgateway --output jsonpath='{.spec.ports[?(@.name=="status-port")].targetPort}')

echo "Determined ingress external address: ${ingress_external_address} and external status port: ${ingress_external_status_port}"

echo "Waiting until it is possible to connect to the ingress gateway"
trial=1
# check if it is possible to establish connection to the ingress gateway (the exact http status code doesn't matter)
until curl --silent --output /dev/null "http://${ingress_external_address}:${ingress_external_status_port}"
do
  if (( trial >= 60 ))
  then
     echo "Exceeded number of trials while waiting for the ingress gateway, giving up..."
     exit 4
  fi
  echo "Ingress gateway does not respond, trying again..."
  sleep 10
  trial=$((trial + 1))
done
echo "Ingress gateway responded"

echo "Executing tests..."
echo "Executing make target $make_target"
make "$make_target"
echo "Tests finished"
