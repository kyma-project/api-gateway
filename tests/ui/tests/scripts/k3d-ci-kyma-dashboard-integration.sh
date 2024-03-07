#!/usr/bin/env bash

set -ex
export CYPRESS_DOMAIN=http://localhost:3001
export DASHBOARD_IMAGE="europe-docker.pkg.dev/kyma-project/prod/kyma-dashboard-local-prod:latest"

sudo apt-get update -y
sudo apt-get install -y gettext-base

function deploy_k3d (){
echo "Provisioning k3d cluster"
sudo k3d cluster create kyma --port 80:80@loadbalancer --port 443:443@loadbalancer --k3s-arg "--disable=traefik@server:0"

export KUBECONFIG=$(k3d kubeconfig merge kyma)

make create-namespace

echo "Apply istio"
kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml

echo "Apply api-gateway"
make deploy
kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/apigateway-default-cr.yaml

echo "Apply gardener resources"
echo "Certificates"
kubectl apply -f https://raw.githubusercontent.com/gardener/cert-management/master/pkg/apis/cert/crds/cert.gardener.cloud_certificates.yaml
echo "DNS Providers"
kubectl apply -f https://raw.githubusercontent.com/gardener/external-dns-management/master/pkg/apis/dns/crds/dns.gardener.cloud_dnsproviders.yaml
echo "DNS Entries"
kubectl apply -f https://raw.githubusercontent.com/gardener/external-dns-management/master/pkg/apis/dns/crds/dns.gardener.cloud_dnsentries.yaml
echo "Issuers"
kubectl apply -f https://raw.githubusercontent.com/gardener/cert-management/master/pkg/apis/cert/crds/cert.gardener.cloud_issuers.yaml

cp $KUBECONFIG tests/ui/tests/fixtures/kubeconfig.yaml
}

function build_and_run_busola() {
echo "Create k3d registry..."
k3d registry create registry.localhost --port=5000

echo "Running kyma-dashboard..."
docker run -d --rm --net=host --pid=host --name kyma-dashboard "$DASHBOARD_IMAGE"

echo "Waiting for the server to be up..."
while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' "$CYPRESS_DOMAIN")" != "200" ]]; do sleep 5; done
sleep 10
}

echo 'Waiting for deploy_k3d_kyma and build_and_run_busola'
deploy_k3d
echo "First process finished"
build_and_run_busola
echo "Second process finished"

cd tests/ui/tests
npm ci && npm run "test:ci"
