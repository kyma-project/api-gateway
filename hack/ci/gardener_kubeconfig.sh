cat <<EOF > gardener_kubeconfig.yaml
apiVersion: v1
kind: Config
current-context: garden-goatz-cli-test
contexts:
  - name: garden-goatz-cli-test
    context:
      cluster: garden
      user: cli-test
      namespace: garden-goatz
clusters:
  - name: garden
    cluster:
      server: https://api.canary.gardener.cloud.sap/
users:
  - name: cli-test
    user:
      token: >-
        $GARDENER_TOKEN
EOF