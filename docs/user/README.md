# API Gateway Module

## What Is API Gateway?

API Gateway provides functionalities that allow you to expose and secure APIs by using [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and the [Istio Service Mesh](https://istio.io/) resources.

## Kyma API Gateway Operator

Kyma API Gateway Operator is an extension to the Kyma runtime that manages the application of API Gateway's configuration and handles resource reconciliation. Within Kyma API Gateway Operator, there are two controllers: [APIGateway Controller](./00-10-overview-api-gateway-controller.md), which applies the configuration specified in the [APIGateway custom resource (CR)](./custom-resources/apigateway/), and [APIRule Controller](./00-20-overview-api-rule-controller.md), which applies the configuration specified in the [APIRule CR](./custom-resources/apirule/).


![Kyma API Gateway Operator Overview](../assets/operator-overview.svg)

## Prerequisites

You must add the Istio module to be able to use the API Gateway module.

## Useful Links

To learn how to use the API Gateway module, read the documentation in the [`user`](../user/) directory. It contains:
- Overview documentation of [APIGateway Controller](./00-10-overview-api-gateway-controller.md) and [APIRule Controller](./00-20-overview-api-rule-controller.md)
- [Tutorials](./tutorials/) that provide step-by-step instructions on creating, exposing, and securing workloads
- Documentation on [APIRule and APIGateway CRs](./custom-resources/)
- Other [technical reference documentation](./technical-reference/)

If you are interested in the detailed documentation of Kyma API Gateway Operator's design and technical aspects, check the [`contributor`](https://github.com/kyma-project/api-gateway/tree/main/docs/contributor) directory.