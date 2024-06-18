# APIRule Controller

## Overview

APIRule Controller is part of Kyma API Gateway Operator. It uses [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and [Istio Service Mesh](https://istio.io/) resources to expose and secure APIs.

## APIRule Custom Resource

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the APIRule custom resource (CR) that is used to expose and secure APIs. See the specification of the APIRule CR in version [`v1beta1`](./custom-resources/apirule/04-10-apirule-custom-resource.md) and [`v2alpha1`](./custom-resources/apirule/v2alpha1/04-10-apirule-custom-resource.md).

## api-gateway-config ConfigMap

The `api-gateway-config` ConfigMap contains the configuration of the **jwt** access strategy.

## Status Codes

The APIRule CR includes status information for all created sub-resources. However, the field **apiRuleStatus** specifically reflects the status of the controller's reconciliation.

| Code          | Description                               |
|---------------|-------------------------------------------|
| **OK**        | Controller finished reconciliation.       |
| **SKIPPED**   | Controller skipped reconciliation.        |
| **ERROR**     | An error occurred during reconciliation.  |


## Controller Limitations

APIRule Controller relies on Istio and Ory custom resources to provide routing capabilities. In terms of persistence, the controller depends only on APIRules stored in the Kubernetes cluster.
In terms of the resource configuration, the following requirements are set on APIGateway Controller:

|          | CPU  | Memory |
|----------|------|--------|
| Limits   | 100m | 128Mi  |
| Requests | 10m  | 64Mi   |

The number of APIRules you can create is not limited.
