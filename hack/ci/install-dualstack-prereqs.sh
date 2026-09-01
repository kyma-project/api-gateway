#!/usr/bin/env bash
#
# Install the -experimental istio-manager needed by the tests/e2e/tests/*
# dualstack suites. The regular manager hard-codes `IsDualStackEnabled=false`
# via the `!experimental` build tag (internal/clusterconfig/regular.go), so it
# never configures the ingressgateway with ipFamilies=[IPv4, IPv6] even when
# kyma-provisioning-info says dualStackIPEnabled: true.
#
# The kyma-system namespace label and the kyma-provisioning-info ConfigMap are
# created by `make create-provisioning-info` (a prerequisite of the
# install-dualstack-prereqs target that invokes this script).
#
# Idempotent. Safe to re-run.
#
# Uses whatever KUBECONFIG the caller has set.

set -euo pipefail
# this installation of istio-manager is only needed for short term, will be removed after finishing the dualstack  https://github.com/kyma-project/istio/issues/2201
ISTIO_MANAGER_VERSION="${ISTIO_MANAGER_VERSION:-latest}"
if [[ "$ISTIO_MANAGER_VERSION" == "latest" ]]; then
  ISTIO_MANAGER_URL="https://github.com/kyma-project/istio/releases/latest/download/istio-manager-experimental.yaml"
else
  ISTIO_MANAGER_URL="https://github.com/kyma-project/istio/releases/download/${ISTIO_MANAGER_VERSION}/istio-manager-experimental.yaml"
fi

tmp="$(mktemp -t istio-manager-experimental.XXXXXX.yaml)"
trap 'rm -f "$tmp"' EXIT
echo "Fetching istio-manager (${ISTIO_MANAGER_VERSION}) from ${ISTIO_MANAGER_URL}"
curl -fsSL --retry 3 --retry-delay 5 -o "$tmp" "$ISTIO_MANAGER_URL"
kubectl apply -f "$tmp"
