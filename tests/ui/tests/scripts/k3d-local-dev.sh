#!/usr/bin/env bash

set -ex

echo "Provisioning k3d cluster"
kyma provision k3d

export KUBECONFIG=$(k3d kubeconfig merge kyma)

echo "Apply istio"
kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml

echo "Apply api-gateway"
kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml
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

cp $KUBECONFIG fixtures/kubeconfig.yaml

