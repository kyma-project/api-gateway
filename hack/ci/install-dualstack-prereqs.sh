#!/usr/bin/env bash
#
# Install the prerequisites the tests/e2e/tests/* suites need to
# exercise dualstack on a Gardener AWS shoot with
# `dualStack.enabled=true` and `ipFamilies=[IPv4, IPv6]`:
#
#   1. Label the kyma-system namespace for istio sidecar injection.
#   2. Apply the `kyma-provisioning-info` ConfigMap that
#      istio-controller-manager reads to gate dualstack
#      (internal/clusterconfig/clusterconfig.go IsDualStackEnabled).
#      On real BTP clusters this is written by infrastructure-manager;
#      Gardener-native CI shoots (and local wet-runs) have to fill it
#      in themselves.
#   3. Install the -experimental istio-manager. The regular manager
#      hard-codes `IsDualStackEnabled=false` via the `!experimental`
#      build tag (internal/clusterconfig/regular.go), so it never
#      configures the ingressgateway with ipFamilies=[IPv4, IPv6] even
#      when kyma-provisioning-info says dualStackIPEnabled: true.
#
# Idempotent. Safe to re-run.
#
# Uses whatever KUBECONFIG the caller has set.

set -euo pipefail

kubectl create namespace kyma-system --dry-run=client -o yaml | kubectl apply -f -
kubectl label namespace kyma-system istio-injection=enabled --overwrite

printf 'networkDetails:\n  dualStackIPEnabled: true\n' \
  | kubectl create configmap -n kyma-system kyma-provisioning-info \
      --from-file=details=/dev/stdin \
      --dry-run=client -o yaml \
  | kubectl apply -f -

kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager-experimental.yaml
