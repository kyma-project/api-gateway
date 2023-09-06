# API Rule Controller

## Overview

API Rule Controller is part of Kyma API Gateway Operator. It uses [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and [Istio Service Mesh](https://istio.io/) resources to expose and secure APIs.

## APIRule CR

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the APIRule CR that is used to expose and secure APIs. To learn more, read the [APIRule CR documentation](./03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md).

## api-gateway-config ConfigMap

The `api-gateway-config` ConfigMap contains the configuration of the JWT Handler. To learn more about how to enable the JWT handling by Istio, read the [APIRule CR documentation](./03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md).

## Status codes

The APIRule CR includes status information for all created sub-resources. However, the field **apiRuleStatus** specifically reflects the status of the controller reconciliation. For more information, read the [APIRule CR documentation](./03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md):

| Code          | Description                               |
|---------------|-------------------------------------------|
| **OK**        | Controller finished reconciliation.       |
| **SKIPPED**   | Controller skipped reconciliation.        |
| **ERROR**     | An error occurred during reconciliation.  |
