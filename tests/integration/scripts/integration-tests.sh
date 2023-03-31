#!/usr/bin/env bash

set -e

function api-gateway::prepare_components_file() {
  log::info "Preparing Kyma installation with Ory and API-Gateway"

cat << EOF > "$PWD/components.yaml"
defaultNamespace: kyma-system
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
  - name: "certificates"
    namespace: "istio-system"
components:
  - name: "istio-resources"
  - name: "ory"
  - name: "api-gateway"
EOF
}

function api-gateway::prepare_test_env_integration_tests() {
  log::info "Prepare test environment variables for integration tests"
  export KYMA_DOMAIN="${CLUSTER_NAME}.${GARDENER_KYMA_PROW_PROJECT_NAME}.shoot.live.k8s-hana.ondemand.com"
}

function api-gateway::launch_integration_tests() {
  log::info "Running API-Gateway integration tests"
  pushd "${API_GATEWAY_SOURCES_DIR}"
#  make install-kyma
  make test-integration
  popd

  log::success "Tests completed"
}

function istio::get_version() {
  pushd "${KYMA_SOURCES_DIR}"
  istio_version=$(git show "${KYMA_VERSION}:resources/istio/Chart.yaml" | grep appVersion | sed -n "s/appVersion: //p")
  popd
}
