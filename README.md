[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/examples)](https://api.reuse.software/info/github.com/kyma-project/examples)

# API Gateway

## What is API Gateway?

API Gateway provides functionalities that allow you to expose and secure APIs by using [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and the [Istio service mesh](https://istio.io/) resources.

## Kyma API Gateway Operator

Kyma API Gateway Operator is an extension to the Kyma runtime that manages the application of API Gateway's configuration and handles resource reconciliation. Within Kyma API Gateway Operator, there are two controllers: [APIGateway Controller](./docs/user/00-10-overview-api-gateway-controller.md), which applies the configuration specified in [APIGateway custom resource (CR)](./docs/user/custom-resources/apigateway/), and [APIRule Controller](./docs/user/00-20-overview-api-rule-controller.md), which applies the configuration specified in [APIRule CR](./docs/user/custom-resources/apirule/).

![Kyma API Gateway Operator Overview](./docs/assets/operator-overview.svg)

## Installation

### Prerequisites

To use API Gateway, you must install Istio and Ory Oathkeeper in your cluster. Learn more about the [API Gateway's dependencies](./docs/contributor/01-20-api-gateway-dependencies.md) and [APIrules' dependencies](./docs/contributor/01-30-api-rule-dependencies.md).

### Procedure
1. Create the `kyma-system` namespace:

    ```bash
    kubectl create namespace kyma-system
    ```

2. To install API Gateway, you must install the latest version of Kyma API Gateway Operator and API Gateway CustomResourceDefinition first. Run:

   ```bash
   kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml
   ```

3. Apply the default API Gateway custom resource (CR):

   ```bash
   kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/apigateway-default-cr.yaml
   ```

   You should get a result similar to this example:

   ```bash
   apigateways.operator.kyma-project.io/default created
   ```

4. Check the state of API Gateway CR to verify if API Gateway was installed successfully:

   ```bash
   kubectl get apigateways/default
   ```

   After successful installation, you get the following output:

   ```bash
   NAME      STATE
   default   Ready
   ```

For more installation options, visit the [installation guide](./docs/contributor/01-00-installation.md).

## Useful links

To learn how to use the API Gateway module, read the documentation in the [`user`](./docs/user/) directory.

If you are interested in the detailed documentation of the Kyma API Gateway Operator's design and technical aspects, check the [`contributor`](./docs/contributor/) directory.

## Contributing

To contribute to this project, follow the general [contributing](https://github.com/kyma-project/community/blob/main/docs/contributing/02-contributing.md) guidelines.
