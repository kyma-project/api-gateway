# API Gateway Module

## What Is API Gateway?

API Gateway is a Kyma module, which provides functionalities that allow you to expose and secure APIs by using [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and [Istio](https://istio.io/) resources.

By default, the API Gateway module is automatically added once you create a Kyma runtime instance. To use the API Gateway module, the Istio module must be also added.

## Features
{TBD}

## Architecture

![Kyma API Gateway Operator Overview](../assets/operator-overview.svg)

### API Gateway Operator

Within the API Gateway moule, API Gateway Operator manages the application of API Gateway's configuration and handles resource reconciliation. It contains two controllers: [APIGateway Controller](./00-10-overview-api-gateway-controller.md), which applies the configuration specified in the [APIGateway custom resource (CR)](./custom-resources/apigateway/), and [APIRule Controller](./00-20-overview-api-rule-controller.md), which applies the configuration specified in the [APIRule CR](./custom-resources/apirule/).


### APIGateway Controller

APIGateway Controller manages the installation of [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and handles the configuration of Kyma Gateway and the resources defined in the APIGateway custom resource (CR). The controller is responsible for:
- Installing, upgrading, and uninstalling Ory Oathkeeper
- Configuring Kyma Gateway
- Managing Certificate and DNSEntry resources
- Configuring Istio Gateways

### APIRule Controller

APIRule Controller uses [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and [Istio](https://istio.io/) resources to expose and secure APIs.

### RateLimit Controller

RateLimit Controller streamlines the implementation of rate limiting, providing means to effectively manage service traffic. It translates configuration specified in RateLimit CR into EnvoyFilters, which Istio uses to enforce rate limiting at the Envoy proxy level. 


## API/Custom Resource Definitions

The apigateways.operator.kyma-project.io CustomResourceDefinition (CRD) describes the APIGateway CR that APIGateway Controller uses to manage the module and its resources. See [APIGateway Custom Resource](./custom-resources/apigateway/04-00-apigateway-custom-resource.md).

The apirules.operator.kyma-project.io CRD describes the APIRule CR that APIRule Controller uses to expose and secure APIs. See [APIRule Custom Resource](./custom-resources/apirule/README.md).

The ratelimits.operator.kyma-project.io CRD describes the RateLimit CR that RateLimit Controller uses to configure rate limiting functionality on the service mesh level. See [APIRule Custom Resource](./custom-resources/apirule/README.md).

## Resource Consumption

To learn more about the resources used by the Istio module, see [Kyma Modules' Sizing](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/kyma-modules-sizing?locale=en-US&state=DRAFT&version=Internal&comment_id=22217515&show_comments=true#api-gateway).