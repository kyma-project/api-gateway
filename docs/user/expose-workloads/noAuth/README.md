# **noAuth** Access Strategy
Use the **noAuth** access strategy to simply expose a workload with no authorization or authentication configurations.

## Configuring **noAuth**

The **noAuth** access strategy provides a simple configuration for exposing workloads. Use it when you need to allow access to specific HTTP methods without any authentication or authorization checks. This setup is suitable for development and testing environments where security requirements are lower and quick access to services is necessary, or when the data being accessed is not sensitive and does not require strict security measures.

See the following sample configuration:
```yaml
...
rules:
  - path: /headers
    methods: ["GET"]
    noAuth: true
```

## **noAuth** Request Flow
The following diagram shows how the **noAuth** access strategy exposes a workload.

![Kyma API Gateway Operator Overview](../../../assets/apirule-noauth.drawio.svg)

To expose a workload with an APIRule and **noAuth**, you need:
- A Kyma Gateway that configures the Istio Ingress Gateway. You can use the default Kyma Gateway or define your own in any namespace. For details, see [Istio Gateways](../../istio-gateways/README.md).
- An APIRule with the **noAuth** access strategy that references:
  - The Service you want to expose.
  - The Istio Gateway (in this case, Kyma Gateway) to route traffic through.

With this setup, a request is processed as follows:
1. A client sends an HTTP request to the exposed hostname, which enters the cluster's Istio Ingress Gateway.
2. Istio Ingress Gateway routs the request straight to the Service based on the APIRule configuration. It doesn't perform any authentication or authorization checks.