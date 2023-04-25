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

## Istio JWT Access Stretegy

This table lists all the possible parameters of the Istio JWT access strategy together with their descriptions:

| Field                                                                     | Description                                                            |
|:--------------------------------------------------------------------------|:-----------------------------------------------------------------------|
| **spec.rules.accessStrategies.config.authentications**                    | List of authentication objects.                                        |
| **spec.rules.accessStrategies.config.authentications.issuer**             | Identifies the issuer that issued the JWT.                             |
| **spec.rules.accessStrategies.config.authentications.jwksUri**            | URL of the provider’s public key set to validate signature of the JWT. |
| **spec.rules.accessStrategies.config.authentications.fromHeaders**        | List of headers from which the JWT token is taken.             |
| **spec.rules.accessStrategies.config.authentications.fromHeaders.name**   | Name of the header.                                                    |
| **spec.rules.accessStrategies.config.authentications.fromHeaders.prefix** | Prefix used before the JWT header. The default is `Bearer`.                    |
| **spec.rules.accessStrategies.config.authentications.fromParams**         | List of parameters from where the JWT token to be taken from.          |
| **spec.rules.accessStrategies.config.authorizations**                     | List of authorization objects.                                         |
| **spec.rules.accessStrategies.config.authorizations.requiredScopes**      | List of required scope values for the JWT.                             |
| **spec.rules.accessStrategies.config.authorizations.audiences**           | List of audiences required for the JWT.                                |

>**NOTE:** Currently we support only a single `fromHeader` or a single `fromParameter`, mixture of both for a JWT issuer is also not supported.

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
              fromHeaders:
              - name: x-jwt-assertion
                prefix: "Kyma "
            - issuer: $ISSUER2
              jwksUri: $JWKS_URI2
              fromParameters:
              - "jwt_token"
            authorizations:
            # Allow only JWTs with the claim "scp", "scope" or "scopes" with the value "test" and the audience "example.com" and "example.org"
            # or JWTs with the claim "scp", "scope" or "scopes" with the values "read" and "write"
            - requiredScopes: ["test"]
              audiences: ["example.com", "example.org"]
            - requiredScopes: ["read", "write"]
```

### Authentications
Under the hood, an authentications array creates a corresponding [requestPrincipals](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Source) array in the Istio Authorization Policy resource.
Every `requestPrincipals` string is formatted as `<ISSUSER>/*`.

### Authorizations
The authorizations field is optional. When not defined, the authorization is satisfied if the JWT is valid. 
You can define multiple authorizations for an access strategy. When multiple authorizations are defined, the request is allowed if at least one of them is satisfied.

The requiredScopes and audiences fields are optional. If requiredScopes is defined, the JWT has to contain all the scopes defined in the requiredScopes field in the `scp`, `scope` or `scopes` claim in order to be authorized.
If audiences is defined, the JWT has to contain all the audiences defined in the audiences field in the `aud` claim in order to be authorized.

## Mutators
For backward compatibility reasons, different types of mutators are supported depending on the access strategy.

| Access Strategy      | Mutator support                                                     |
|:---------------------|:--------------------------------------------------------------------|
| jwt                  | Istio-based cookie and header mutator                               |
| oauth2_introspection | [Ory mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) |
| noop                 | [Ory mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) |
| allow                | No mutators supported                                               |

### Istio-based Mutators
Mutators can be used to enrich an incoming request with information. The following mutators are supported in combination with 
the `jwt` access strategy and can be defined for each rule in an `ApiRule`: `header`,`cookie`. 
It's possible to configure multiple mutators for one rule, but only one mutator of each type is allowed.

#### Header Mutator
The headers are specified via the `headers` field of the header mutator configuration field. The keys are the names of the headers and the values are a string.
In the header value it is possible to use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators), 
e.g. to write an incoming header to a new header.
The configured headers are set to the request and overwrite all existing headers with the same name.

See the example:
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
      mutators:
        - handler: header
          config:
            headers:
              # Add a new header called X-Custom-Auth with the value of the incoming Authorization header
              X-Custom-Auth: "%REQ(Authorization)%"
              # Add a new header called X-Some-Data with the value "some-data"
              X-Some-Data: "some-data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```

#### Cookie Mutator
The cookies are specified via the `cookies` field of the cookie mutator configuration field. The keys are the names of the cookies and the values are a string. 
In the cookie value it is possible to use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators).
The configured cookies are set as `Cookie`-header in the request and overwrite an existing `Cookie`-header.

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
      mutators:
        - handler: cookie
          config:
            cookies:
              # Add a new cookie called some-data with the value "data"
              some-data: "data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```
