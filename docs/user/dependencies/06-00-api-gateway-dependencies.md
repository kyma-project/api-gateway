# API Gateway module depenedencies

## Istio

API Gateway requires Istio installed on the cluster. This is required as API Gateway creates the custom resources `Gateway` and `VirtualService` provided by Istio. The recommended installation option for Istio is the [Kyma Istio Operator](https://github.com/kyma-project/istio#install-kyma-istio-operator-and-istio-from-the-latest-release).

## Dependecies in BTP Kyma Runtime

Additionally in `BTP Kyma Runtime` API Gateway uses `DNSEntry` and `Certificate` Custom Resources provided by [Gardener](https://gardener.cloud). The resources should be present in the Kyma instance with no additional steps needed, as `BTP Kyma Runtime` uses `Gardener` cluster as the running environment.