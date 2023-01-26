# API-Gateway JWT Handler Feature Toggle

## Overview

We support two JWT handlers in API-Gateway at the moment. Switching between them is configurable via the following ConfigMap: `kyma-system/api-gateway-config`

## Switching between JWT handlers

### Enabling `ory/oathkeeper` JWT handler

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: ory"}}'
```

### Enabling `istio` JWT handler

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: istio"}}'
```

# Using Istio JWT Handler

When `istio` JWT Handler is enabled you can configure APIRule with Istio JWT like in the example below:

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
  namespace: $NAMESPACE
spec:
  gateway: kyma-system/kyma-gateway
  host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: $ISSUER
              jwksUri: $JWKS_URI
            authorizations:
            - requiredScopes: ["test"]
            - requiredScopes: ["read", "write"]
```

The `authorizations` field defined above will require your JWT to contain either `test` scope OR `read` AND `write` scopes. The scope value has to be in one of the following keys: `scp`, `scope`, `scopes`, to ensure backwards compatibility with ory oathkeeper.