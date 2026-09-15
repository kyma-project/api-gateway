#!/usr/bin/env bash

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars GARDENER_TOKEN

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
