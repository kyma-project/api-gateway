#!/usr/bin/env bash

if [[ -n "$LOCAL_IMAGE" ]]; then
  IMG=$LOCAL_IMAGE
fi


if [[ -z "$IMG" ]]; then
  echo "Please specify APIGateway image to use during the test"
  echo "For most cases, you can use the latest released image from the APIGateway repository, for example:"
  echo "export IMG=europe-docker.pkg.dev/kyma-project/prod/api-gateway/releases/api-gateway-manager:\{RELEASE_VERSION\}"
  echo "You can also use locally built image, by setting LOCAL_IMAGE variable, for example:"
  echo "export LOCAL_IMAGE=api-gateway-manager:latest"
  exit 1
fi

echo "Running tests on prod Busola"

set -e
export CYPRESS_DOMAIN=http://localhost:3001
export DASHBOARD_IMAGE="europe-docker.pkg.dev/kyma-project/prod/kyma-dashboard-local-prod:latest"

function deploy_k3d (){
echo "Provisioning k3d cluster"

echo "Will remove existing k3d cluster and registry"
echo "Proceed with deletion (Y/n)?"


read -r response

if [[ "$response" != "n" ]]; then
  k3d cluster delete kyma || true
  k3d registry delete registry.localhost || true
fi

k3d cluster create kyma --port 80:80@loadbalancer --port 443:443@loadbalancer --k3s-arg "--disable=traefik@server:0" --k3s-arg '--tls-san=host.docker.internal@server:*' --image 'rancher/k3s:v1.31.7-k3s1'

if [[ -n "$LOCAL_IMAGE" ]]; then
  k3d image import "$IMG" -c kyma
fi

export KUBECONFIG=$(k3d kubeconfig merge kyma)

kubectl create ns kyma-system

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

kubectl wait --for=jsonpath='{.status.state}'=Ready --timeout=300s apigateway default

cp $KUBECONFIG tests/ui/tests/fixtures/kubeconfig.yaml
}

function build_and_run_busola() {
echo "Create k3d registry..."
k3d registry create registry.localhost

docker kill kyma-dashboard || true
echo "Running kyma-dashboard with image $DASHBOARD_IMAGE..."
docker run -d --rm -e DOCKER_DESKTOP_CLUSTER=true --env ENVIRONMENT=prod -p 3001:3001 --name kyma-dashboard "$DASHBOARD_IMAGE"

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
npm ci && npm run "test"
