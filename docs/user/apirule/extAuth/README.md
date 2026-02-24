# Configure the **extAuth** Access Strategy

**extAuth** is an access strategy allowing for providing custom authentication and authorization logic. To use it, you must first define the authorization provider in the Istio configuration, most commonly in the Istio custom resource (CR). For example, the following Istio CR defines a provider named `ext-auth-provider`.

```yaml
apiVersion: operator.kyma-project.io/v1alpha2
kind: Istio
metadata:
  name: default
  namespace: kyma-system
spec:
  config:
    authorizers:
    - headers:
        inCheck:
          include:
          - x-ext-authz
      name: ext-auth-provider
      port: 8000
      service: ext-auth-provider.provider-system.svc.cluster.local
```

Once you define the provider in the Istio configuration, you can reference it in the APIRule CR.

### Securing an Endpoint with **extAuth**
This configuration allows access to the `/get` path of the `user-service` service. Based on this APIRule, a `CUSTOM` Istio AuthorizationPolicy is created with the `ext-auth-provider` provider, securing access.

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: ext-authz
  namespace: user-system
spec:
  hosts:
    - na.example.com
  service:
    name: user-service
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: "/get"
      methods: ["GET"]
      extAuth:
        authorizers:
        - x-ext-authz
```

### Securing an Endpoint with **extAuth** and JWT Restrictions

This configuration allows access to the `/get` path of the `user-service` Service, as in the example in the previous section. The access is further restricted by the JWT configuration, specified in the **restrictions** field.

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: ext-authz
  namespace: user-system
spec:
  hosts:
    - na.example.com
  service:
    name: user-service
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: "/get"
      methods: ["GET"]
      extAuth:
        authorizers:
          - x-ext-authz
        restrictions:
          authentications:
            - issuer: https://example.com
              jwksUri: https://example.com/.well-known/jwks.json
          authorizations:
            - audiences: ["app1"]
```