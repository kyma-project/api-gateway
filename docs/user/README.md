# API Gateway Module

## What Is API Gateway?

API Gateway is a Kyma module with which you can expose and secure APIs.

To use the API Gateway module, you must also add the Istio module. Moreover, to expose a workload using the APIRule custom resource (CR), the workload must be part of the Istio service mesh. 

By default, both the API Gateway and Istio modules are automatically added when you create a Kyma runtime instance. 

## Features

The API Gateway module offers the following features:

- API Exposure: The module uses Istio features to help you easily and securely expose your workloads by creating APIRule CRs. With an APIRule, you can perform the following actions:
  - Group multiple workloads and expose them under a single host.
  - Use a short host name to simplify the migration of resources to a new cluster.
  - Configure the **noAuth** access strategy, which offers a simple configuration to allow access to specific HTTP methods.
  - Secure your workloads by configuring **jwt** or **extAuth** access strategies. The **jwt** access strategy enables you to use Istio's JWT configuration to protect your exposed services and interact with them using JSON Web Tokens. The **extAuth** access strategy allows you to implement custom authentication and authorization logic.

- Default Kyma Gateway configuration: The module manages the default Istio TLS Gateway that handles incoming traffic in your Kyma cluster. The Gateway uses the cluster's default domain and a self-signed certificate.

- Rate Limiting: The module simplifies local rate limiting on the Istio service mesh layer. By configuring a RateLimit CR, you can limit the number of requests targeting an exposed application in a unit of time, based on specific paths and headers.

- External Gateway integration: Configure Istio mTLS Gateway and integrate it with external gateway deployed outside of the Kyma cluster.

## Architecture

Within the API Gateway module, API Gateway Operator manages the application of API Gateway's configuration and handles resource reconciliation. It contains the following controllers: APIGateway Controller, APIRule Controller, RateLimit Controller, and ExternalGateway Controller. See the following diagram:

![Kyma API Gateway Operator Overview](../assets/operator-overview.drawio.svg)

## API/Custom Resource Definitions

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the APIGateway CR that APIGateway Controller uses to manage the module and its resources. See [APIGateway Custom Resource](./custom-resources/apigateway/04-00-apigateway-custom-resource.md).

The `apirules.operator.kyma-project.io` CRD describes the APIRule CR that APIRule Controller uses to expose and secure APIs. See [APIRule Custom Resource](./custom-resources/apirule/04-10-apirule-custom-resource.md).

The `externalgateways.gateway.kyma-project.io` CRD describes the kind and the format of data that ExternalGateway Controller uses to configure external gateway integration. See [ExternalGateway Custom Resource](../user/custom-resources/externalgateway/externalgateway-custom-resource.md).

The `ratelimits.gateway.kyma-project.io` CRD describes the kind and the format of data that RateLimit Controller uses to configure request rate limits for applications. See [RateLimit Custom Resource](./local-rate-limit.md).

## Authorization

To assign access permissions to the API Gateway module resources, use the following [aggregated ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles):

- `kyma-api-gateway-view`- Grants read-only access to all API Gateway resources.
- `kyma-api-gateway-edit` - Grants full access to `gateway.kyma-project.io` resources and read-only access to `operator.kyma-project.io` resources.
- `kyma-api-gateway-admin` - Grants full access to all API Gateway resources.

## Resource Consumption

To learn more about the resources used by the Istio module, see [Kyma Modules' Sizing](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/kyma-modules-sizing?locale=en-US&state=DRAFT&version=Internal&comment_id=22217515&show_comments=true#api-gateway).
