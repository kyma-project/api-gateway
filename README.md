[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/examples)](https://api.reuse.software/info/github.com/kyma-project/examples)

# API Gateway

## What is API Gateway?

API Gateway provides functionalities that allow you to expose and secure APIs by using [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and the [Istio service mesh](https://istio.io/) resources.

## API Gateway Operator

API Gateway Operator is a component of the Kyma runtime that handles resource reconciliation and manages the application of API Gateway configuration. Within API Gateway Operator, there are two controllers: [APIGateway Controller](./00-10-overview-api-gateway-controller.md), which applies the configuration specified in APIGateway CR, and [APIRule Controller](./00-20-overview-api-rule-controller.md), which applies the configuration specified in APIRule CR.

![Kyma API Gateway Operator Overview](./docs/assets/operator-overview.svg)

## Installation

See how to [install API Gateway](./docs/contributor/01-00-installation.md).

## Useful links

To learn how to use Kyma API Gateway, read the documentation in the [`user`](./docs/user/) directory. 

If you are interested in the detailed documentation of Kyma API Gateway Operator's design and technical aspects, check the [`contributor`](./docs/contributor/) directory.

## Contributing

To contribute to this project, follow the general [contributing](https://github.com/kyma-project/community/blob/main/docs/contributing/02-contributing.md) guidelines.