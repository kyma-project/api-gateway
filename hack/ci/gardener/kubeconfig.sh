#!/usr/bin/env bash

set -eo pipefail

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
    GARDENER_TOKEN
)

check_required_vars "${requiredVars[@]}"

cat <<EOF > gardener_kubeconfig.yaml
apiVersion: v1
kind: Config
current-context: garden-goats-github
contexts:
  - name: garden-goats-github
    context:
      cluster: garden
      user: github
      namespace: garden-goats
clusters:
  - name: garden
    cluster:
      server: https://api.live.gardener.cloud.sap
users:
  - name: github
    user:
      token: >-
        $GARDENER_TOKEN
EOF