[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/api-gateway)](https://api.reuse.software/info/github.com/kyma-project/api-gateway)

# API Gateway

API Gateway is a Kyma module with which you can expose and secure APIs.

The API Gateway module offers the following features:

- Ory Oathkeeper installation: The module simplifies and manages the installation of Ory Oathkeeper.
- API Exposure: The module combines ORY Oathkeeper and Istio capabilities to offer the APIRule CustomResourceDefinition. By creating APIRule custom resources, you can easily and securely expose your workloads.
- Kyma Gateway installation: The module installs the default simple TLS Kyma Gateway.

## Installation

### Prerequisites

To use the API Gateway module, you must install Istio and Ory Oathkeeper in your cluster. Learn more about the [API Gateway's dependencies](./docs/contributor/01-20-api-gateway-dependencies.md) and [APIrules' dependencies](./docs/contributor/01-30-api-rule-dependencies.md).

### Procedure
1. Create the `kyma-system` namespace and label it with `istio-injection=enabled`:

   ```bash
   kubectl create namespace kyma-system
   kubectl label namespace kyma-system istio-injection=enabled --overwrite
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

For more installation options, see the [installation guide](./docs/contributor/01-00-installation.md).

## Useful Links

To learn how to use the API Gateway module, read the documentation in the [`user`](./docs/user/) directory.

If you are interested in the detailed documentation of the Kyma API Gateway Operator's design and technical aspects, check the [`contributor`](./docs/contributor/) directory.

## Contributing
<!--- mandatory section - do not change this! --->

See the [Contributing](CONTRIBUTING.md) guidelines.

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
