# API Gateway CRD proposal

This document describes proposed API for installing APIGateway component.

## Proposed CR structure

```yaml
kind: APIGateway
spec:
  enableKymaGateway: true # Part of the custom resource gby default (or true if not defined)
  gateways:
    - namespace: "some-ns" # Required
      name: "gateway1" # Required
      servers:
        - hosts: # Creating  more than one for the same host:port configuration should result in  Warning
            - host: "goat.example.com"
              certificate: "goat-certificate" # If not defined, generate Gardener certificate
            - host: "goat1.example.com"
              dnsProviderSecret: "my-namespace/dns-secret" # If provided generate a DNS Entry with Gardener 
          port:
            number: 443
            name: https
            protocol: HTTPS
            TLS: MUTUAL
        - hosts:
            - host: "*.goat.example.com"
              default: true # Use as deafult host for API Rules
            - host: "goat1.example.com"
              dnsProviderSecret: "my-namespace/dns-secret" # If provided generate a DNS Entry with Gardener 
          port:
            number: 80
            name: http
            protocol: HTTP
          httpsRedirect: true # If on Protocol = HTTPS, set Warning
status:
  state: "Warning"
  description: "Cannot have same host on two gateways"
  conditions:
  - '[...]' # array of *metav1.Condition

```

## Example use cases

### User wants to use api-gateway with no additional configuration

User creates an APIGateway CR with no additional configuration.

```yaml
kind: APIGateway
namespace: kyma-system
name: default
```

By default, APIGateway will generate a Certificate and DNSEntry for default Kyma domain. With this configuration user can expose their workloads under Kyma domain.

### A managed Kyma user wants to expose their workloads under custom domain

Prerequisite:
- A DNS secret `dns-secret` exists in namespace `my-namespace`

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

As the cluster is managed Kyma cluster (SKR), a DNSProvider with the provided secret and a DNSEntry will be created. If the user does not configure port type then the Istio Gateway will be generated with both HTTP and HTTPS by default. Additional Gardener Certificate will also be created and provided for HTTPS.

The user can now already expose their services under `test.example.com` and `test2.example.com`.

### User wants to expose their Mongo instance

User configures API Gateway as follows: 

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

This allows access to Mongo DB under host `mongo.example.com:2379`.
