# APIGateway Controller

## Overview

APIGateway Controller is part of Kyma API Gateway Operator. Its role is to manage the installation of [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and handle the configuration of Kyma Gateway and the resources defined by the [APIGateway custom resource (CR)](./custom-resources/apigateway/04-00-apigateway-custom-resource.md). The controller is responsible for:
- Installing, upgrading, and uninstalling Ory Oathkeeper
- Configuring Kyma Gateway
- Managing Certificates and DNS entries
- Configuring Istio Gateways

## APIGateway CR

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the APIGateway CR that is used to manage the API Gateway resources. To learn more, read the [APIGateway CR documentation](./custom-resources/apigateway/04-00-apigateway-custom-resource.md).

## Status codes

|     Code     | Description                              |
|:------------:|:-----------------------------------------|
|   `Ready`    | Controller finished reconciliation.      |
| `Processing` | Controller is reconciling resources.     |
|  `Deleting`  | Controller is deleting resources.        |
|   `Error`    | An error occurred during reconciliation. |
|  `Warning`   | Controller is misconfigured.             |