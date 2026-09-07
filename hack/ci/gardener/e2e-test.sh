#!/usr/bin/env bash

# Description: This script runs given integration tests on a real Gardener cluster
# It installs istio and api gateway and then runs make test targets provided via commandline arguments to that script
# It requires the following env variables:
# - IMG - API gateway image to be deployed (by make deploy)
# - CLUSTER_NAME - Gardener cluster name
# - CLUSTER_KUBECONFIG - Gardener cluster kubeconfig path
# - GARDENER_CONFIGURATION - configuration preset; provides GARDENER_IP_STACK
#   via configurations/${GARDENER_CONFIGURATION}/vars.sh. When the shoot is
#   dualstack (GARDENER_IP_STACK=dualstack) the experimental istio-manager is
#   installed (the only build that honours dualStackIPEnabled).
# Optional:
# - TEST_IP_FAMILY - ipv4 (default) | ipv6 | dualstack. Selects only the test
#   client dial family; it does NOT drive infra dualstack (that is the shoot's
#   GARDENER_IP_STACK, above).

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_positional make_target "$1"

require_vars IMG CLUSTER_NAME CLUSTER_KUBECONFIG GARDENER_CONFIGURATION

load_configuration "${GARDENER_CONFIGURATION}"

require_vars GARDENER_IP_STACK

echo "Make target: ${make_target}"

echo "Executing tests in cluster ${CLUSTER_NAME}, kubeconfig ${CLUSTER_KUBECONFIG}"
export KUBECONFIG="${CLUSTER_KUBECONFIG}"

export CLUSTER_DOMAIN=$(kubectl get configmap -n kube-system shoot-info -o jsonpath="{.data.domain}")
echo "Cluster domain: ${CLUSTER_DOMAIN}"

export GARDENER_PROVIDER=$(kubectl get configmap -n kube-system shoot-info -o jsonpath="{.data.provider}")
echo "Gardener provider: ${GARDENER_PROVIDER}"

export TEST_DOMAIN="${CLUSTER_DOMAIN}"
export IS_GARDENER=true # this variable is used in tests to make decisions based on the fact that the tests are running in Gardener

[[ "${GARDENER_IP_STACK}" == "dualstack" ]] && DUAL_STACK_ENABLED="true" || DUAL_STACK_ENABLED="false"
echo "Dual stack enabled: ${DUAL_STACK_ENABLED}"

# Add pwd to path to be able to use binaries downloaded in scripts
export PATH="${PATH}:${PWD}"

start_group "Creating kyma-system namespace"
make create-namespace
end_group

start_group "Creating kyma-provisioning-info configmap"
make create-provisioning-info DUAL_STACK_ENABLED="${DUAL_STACK_ENABLED}"
end_group

start_group "Installing istio"
if [ "${GARDENER_IP_STACK:-}" = "dualstack" ]; then
  make install-istio-experimental
else
  make install-istio
fi
end_group

start_group "Deploying api-gateway, image: ${IMG}"
make deploy
end_group

start_group "Waiting for the ingress gateway external address"
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
end_group

start_group "Executing tests: ${make_target}"
make "${make_target}"
end_group
