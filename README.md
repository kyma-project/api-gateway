[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/api-gateway)](https://api.reuse.software/info/github.com/kyma-project/api-gateway)

# API Gateway

API Gateway is a Kyma module with which you can expose and secure APIs.

The API Gateway module offers the following features:

- API Exposure: The module uses Istio features to help you easily and securely expose your workloads by creating APIRule custom resources. With an APIRule, you can perform the following actions:
  - Group multiple workloads and expose them under a single host.
  - Use a short host name to simplify the migration of resources to a new cluster.
  - Configure the **noAuth** access strategy, which offers a simple configuration to allow access to specific HTTP methods.
  - Secure your workloads by configuring **jwt** or **extAuth** access strategies. The **jwt** access strategy enables you to use Istio's JWT configuration to protect your exposed services and interact with them using JSON Web Tokens. The **extAuth** access strategy allows you to implement custom authentication and authorization logic.

- Gateway configuration:
  - Default Kyma Gateway: The module sets up the default TLS Kyma Gateway, which uses the default domain and a self-signed certificate.
  - Custom Gateway: The module allows you to configure a custom Gateway, which is recommended for production environments. Additionally, it enables you to expose workloads using a custom domain and DNSEntry. 

- Rate Limiting: The module simplifies local rate limiting on the Istio service mesh layer. You can configure it using a straightforward RateLimit custom resource.

## Installation

### Prerequisites

- [k3d](https://k3d.io/stable/)

### Procedure

1. Create a Kyma cluster.

      ```bash
      k3d cluster create kyma --port 80:80@loadbalancer --port 443:443@loadbalancer --k3s-arg "--disable=traefik@server:*"
      ```

2. Create the kyma-system namespace with enabled Istio sidecar injection.

      ```bash
      kubectl create ns kyma-system
      kubectl label namespace kyma-system istio-injection=enabled --overwrite
      ```

2. To use the API Gateway module, you must also add the Istio module. Run:

      ```bash
      kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
      kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
      ```

      To verify if the Istio module is added, check the state of the Istio CR:

      ```bash
      kubectl get istios/default -n kyma-system
      ```

      If successful, you get the following output:

      ```bash
      NAME      STATE
      default   Ready
      ```

3. Add the API Gateway module.

      ```bash
      kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml
      kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/apigateway-default-cr.yaml
      ```

      To verify if the API Gateway module is added, check the state of the API Gateway CR:

      ```bash
      kubectl get apigateways/default
      ```

      If successful, you get the following output:

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
