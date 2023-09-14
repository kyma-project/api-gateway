[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/examples)](https://api.reuse.software/info/github.com/kyma-project/examples)

# API Gateway

## What is API Gateway?

API Gateway provides functionalities that allow you to expose and secure APIs by using [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and the [Istio Service Mesh](https://istio.io/) resources.

## Kyma API Gateway Operator

Kyma API Gateway Operator is a component of the Kyma runtime that handles the reconciliation of resources to apply the API Gateway configuration. Within Kyma API Gateway Operator, [API Gateway Controller](./00-10-overview-api-gateway-controller.md) and [API Rule Controller](./00-20-overview-api-rule-controller.md) are responsible for applying this configuration.

![Kyma API Gateway Operator Overview](./docs/assets/operator-overview.svg)

## Installation

See how to [install API Gateway](./docs/contributor/01-00-installation.md).

## Useful links

To learn how to use Kyma API Gateway, read the documentation in the [`user`](./docs/user/) directory. 

If you are interested in the detailed documentation of Kyma API Gateway Operator's design and technical aspects, check the [`contributor`](./docs/contributor/) directory.

## Contributing

To contribute to this project, follow the general [contributing](https://github.com/kyma-project/community/blob/main/docs/contributing/02-contributing.md) guidelines.