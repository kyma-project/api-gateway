# API Gateway

## What is API Gateway

API Gateway provides functionalities that allow you to expose and secure APIs by using [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and the [Istio Service Mesh](https://istio.io/) resources.

## Kyma API Gateway Operator

Kyma API Gateway Operator is a component of the Kyma runtime that handles the reconciliation of resources to apply the API Gateway configuration. Within Kyma API Gateway Operator, [API Gateway Controller](./00-10-overview-api-gateway-controller.md) and [API Rule Controller](./00-20-overview-api-rule-controller.md) are responsible for applying this configuration.

![Kyma API Gateway Operator Overview](../assets/operator-overview.svg)


## Useful links

To learn how to use Kyma API Gateway, read the documentation in the [`user`](../user/) directory. It contains:
- Overview documentation of [APIGateway Controller](./00-10-overview-api-gateway-controller.md) and [APIRule Controller](./00-20-overview-api-rule-controller.md)
- [Tutorials](./01-tutorials/) that provide step-by-step instructions on creating, exposing, and securing workloads
- Technical reference documentation that describe APIRule and APIGateway custom resources, lists API Gateway Operator parameters, and more.

If you are interested in the detailed documentation of Kyma API Gateway Operator's design and technical aspects, check the [`contributor`](../contributor/) directory.