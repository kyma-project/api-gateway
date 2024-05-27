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
      server: https://api.live.gardener.cloud.sap/
users:
  - name: github
    user:
      token: >-
        $GARDENER_TOKEN
EOF