# Exposing and Securing Workloads with APIRules
(#exposing-and-securing-workloads-with-apirules)

- [Exposing and Securing Workloads with APIRules](#exposing-and-securing-workloads-with-apirules)
  - [What Is an APIRule?](#what-is-an-apirule)
  - [APIRule Capabilities](#apirule-capabilities)
    - [Add Authentication and Authorization](#add-authentication-and-authorization)
      - [JWT](#jwt)
      - [extAuth](#extauth)
      - [noAuth](#noauth)
    - [Expose Multiple Workloads on One Host](#expose-multiple-workloads-on-one-host)
    - [Use a Short Host Name](#use-a-short-host-name)


## What Is an APIRule?


## APIRule Capabilities

### Add Authentication and Authorization

#### JWT

The `jwt` access strategy allows you to secure your workload with HTTPS using Istio JWT configuration. It offers a secure and efficient method for protecting your services and interacting with them using JSON Web Tokens (JWTs). This approach is highly recommended if you aim to secure your workloads without the need to implement custom authentication and authorization logic.

See [Expose and Secure a Workload with a JWT Using SAP Cloud Identity Services](./01-40-expose-workload-jwt.md).

> [!NOTE]
> For most common scenarios, it is best to configure the `jwt` access strategy within the APIRule CR.

#### extAuth

    The `extAuth` access strategy allows you to provide your custom authorization and authentication logic. Use this access strategy when the built-in Istio JWT authentication and authorization mechanisms do not meet your specific requirements, and you want to offload the logic to a custom external service. It provides flexibility and the ability to tailor security to your application's specific needs.

    To use the `extAuth` access strategy, you must first define the authorization provider in the Istio configuration, most commonly in the Istio custom resource (CR). Once you define the provider in the Istio configuration, you can reference it in the APIRule CR, specifying which endpoints and methods should be protected by this provider.

    See [Expose and Secure a Workload with OAuth2 Proxy External Authorizer (Client Credentials Flow)](./01-53-expose-workload-extauth-client-credentials.md) and [Expose and Secure a Workload with OAuth2 Proxy External Authorizer (Authorization Code Flow)](./01-53-expose-workload-extauth-client-credentials.md).

#### noAuth

    The `noAuth` access strategy provides a simple configuration for exposing workloads. Use it when you simply need to expose your workloads and allow access to specific HTTP methods without any authentication or authorization checks. This setup is suitable for development and testing environments where security requirements are lower and quick access to services is necessary, or when the data being accessed is not sensitive and does not require strict security measures.

    > [!WARNING]
    > Exposing a workload to the outside world is a potential security vulnerability, so be careful. In a production environment, always secure the workload you expose.

    See [Exposing Workloads with noAuth](./01-35-expose-workload-noauth-gardener.md).

### Expose Multiple Workloads on One Host

You can configure a single APIRule to expose multiple Services on different paths, allowing you to consolidate the routing configuration for related workloads. See [Expose Multiple Workloads](./01-41-expose-multiple-workloads.md).

### Use a Short Host Name
Instead of using the default fully qualified domain name, you can configure a shorter, more user-friendly host name for your APIRule. See [Use a Short Host](./01-43-expose-workload-short-host-name.md).

