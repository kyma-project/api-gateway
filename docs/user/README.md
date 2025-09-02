# API Gateway Module

## What Is API Gateway?

API Gateway is a Kyma module with which you can expose and secure APIs.

To use the API Gateway module, you must also add the Istio module. Moreover, to expose a workload using the APIRule custom resource, the workload must be part of the Istio service mesh. 

By default, both the API Gateway and Istio modules are automatically added when you create a Kyma runtime instance. 

## Features

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

## Architecture

![Kyma API Gateway Operator Overview](../assets/operator-overview.svg)

### API Gateway Operator

Within the API Gateway module, API Gateway Operator manages the application of API Gateway's configuration and handles resource reconciliation. It contains the following controllers: APIGateway Controller, APIRule Controller, and RateLimit Controller.


### APIGateway Controller

APIGateway Controller handles the configuration of Kyma Gateway. The controller is responsible for the following:
- Configuring Kyma Gateway
- Managing Certificate and DNSEntry resources

### APIRule Controller

APIRule Controller uses [Istio](https://istio.io/) resources to expose and secure APIs.

### RateLimit Controller

RateLimit Controller manages the configuration of local rate limiting on the Istio service mesh layer. By creating a RateLimit custom resource (CR), you can limit the number of requests targeting an exposed application in a unit of time, based on specific paths and headers.

## API/Custom Resource Definitions

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the APIGateway CR that APIGateway Controller uses to manage the module and its resources. See [APIGateway Custom Resource](./custom-resources/apigateway/04-00-apigateway-custom-resource.md).

The `apirules.operator.kyma-project.io` CRD describes the APIRule CR that APIRule Controller uses to expose and secure APIs. See [APIRule Custom Resource](./custom-resources/apirule/README.md).

The `ratelimits.gateway.kyma-project.io` CRD describes the kind and the format of data that RateLimit Controller uses to configure request rate limits for applications. See [RateLimit Custom Resource](./custom-resources/ratelimit/04-00-ratelimit.md).

## Resource Consumption

To learn more about the resources used by the Istio module, see [Kyma Modules' Sizing](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/kyma-modules-sizing?locale=en-US&state=DRAFT&version=Internal&comment_id=22217515&show_comments=true#api-gateway).
