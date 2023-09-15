# API Rule Controller

## Overview

API Rule Controller is part of Kyma API Gateway Operator. It uses [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and [Istio Service Mesh](https://istio.io/) resources to expose and secure APIs.

## APIRule CR

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the APIRule CR that is used to expose and secure APIs. To learn more, read the [APIRule CR documentation](./custom-resources/apirule/04-10-apirule-custom-resource.md).

## api-gateway-config ConfigMap

The `api-gateway-config` ConfigMap contains the configuration of the JWT Handler. To learn more about how to [enable Istio the JWT handling](./custom-resources/apirule/04-20-apirule-istio-jwt-access-strategy.md).

## Status codes

The APIRule CR includes status information for all created sub-resources. However, the field **apiRuleStatus** specifically reflects the status of the controller's reconciliation.

| Code          | Description                               |
|---------------|-------------------------------------------|
| **OK**        | Controller finished reconciliation.       |
| **SKIPPED**   | Controller skipped reconciliation.        |
| **ERROR**     | An error occurred during reconciliation.  |


## Controller limitations

APIRule Controller relies on Istio and Ory Custom Resources to provide routing capabilities. In terms of persistence, the controller depends only on APIRules stored in the Kubernetes cluster.
In terms of the resource configuration, the following requirements are set on API Gateway Controller:

|          | CPU  | Memory |
|----------|------|--------|
| Limits   | 100m | 128Mi  |
| Requests | 10m  | 64Mi   |

The number of APIRules you can create is not limited. 