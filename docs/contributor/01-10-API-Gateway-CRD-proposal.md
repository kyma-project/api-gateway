# API Gateway CRD proposal

This document describes the proposed API for installing the APIGateway component.

## Proposed CR structure

```yaml
kind: APIGateway
spec:
  enableKymaGateway: true # Part of the custom resource by default
  defaultHost: "example.com" # Use as default host for API Rules. If not defined and `enableKymaGateway: true`, use the Kyma host. If both fields are false, require a full host in API Gateway
  gateways:
    - namespace: "some-ns" # Required
      name: "gateway1" # Required
      servers:
        - hosts: # Creating  more than one host for the same **host:port** configuration results in a `Warning`
            - host: "goat.example.com"
              certificate: "goat-certificate" # If not defined, generate a Gardener certificate
            - host: "goat1.example.com"
              dnsProviderSecret: "my-namespace/dns-secret" # If provided, generate a DNS Entry with Gardener 
          port:
            number: 443
            name: https
            protocol: HTTPS
            TLS: MUTUAL
        - hosts:
            - host: "*.goat.example.com"
            - host: "goat1.example.com"
              dnsProviderSecret: "my-namespace/dns-secret" # If provided, generate a DNS Entry with Gardener 
          port:
            number: 80
            name: http
            protocol: HTTP
          httpsRedirect: true # If `Protocol = HTTPS`, set `Warning`
status:
  state: "Warning"
  description: "Cannot have the same host for two gateways"
  conditions:
  - '[...]' # array of *metav1.Condition

```

## Example use cases

### The user wants to use API Gateway with no additional configuration

The user creates an APIGateway CR with no additional configuration.

```yaml
kind: APIGateway
namespace: kyma-system
name: default
```

By default, APIGateway generates a Certificate and DNSEntry for the default Kyma domain. With this configuration, the user can expose their workloads under the Kyma domain.

### The managed Kyma user wants to expose their workloads under a custom domain

Prerequisite:
- The DNS secret `dns-secret` exists in the `my-namespace` namespace.

The user configures the CR as follows:

```yaml
kind: APIGateway
namespace: kyma-system
name: default
spec:
  gateways:
    - namespace: "my-namespace"
      name: "test-gateway"
      servers:
        - hosts:
            - host: "test.example.com"
              dnsProviderSecret: "my-namespace/dns-secret"
            - host: "test2.example.com"
              dnsProviderSecret: "my-namespace/dns-secret"
```

Because it is a managed Kyma cluster (SKR), a DNSProvider with the provided Secret and a DNSEntry are created. If the user does not configure the port type, Istio Gateway is generated with both HTTP and HTTPS. An additional Gardener Certificate is also created and provided for HTTPS.

The user can now expose their services under the hosts `test.example.com` and `test2.example.com`.

### The user wants to expose their Mongo instance

The user configures API Gateway as follows: 

```yaml
kind: APIGateway
namespace: kyma-system
name: default
spec:
  gateways:
    - namespace: "my-namespace"
      name: "mongo-gateway"
      servers:
        - hosts:
            - host: "mongo.example.com"
          port:
              number: 2379
              name: mongo
              protocol: MONGO
```

It allows access to Mongo DB under the host `mongo.example.com:2379`.
