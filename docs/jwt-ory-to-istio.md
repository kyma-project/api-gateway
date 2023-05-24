---
title: Differences between Ory Oathkeeper and Istio JWT access strategies
---

We are in the process of transitioning from Ory Oathkeeper to Istio JWT access strategy. This document explains the differences between those two strategies and compares their configuration.

## Comparison of Ory Oathkeeper and Istio JWT access strategy configurations

>**CAUTION:** Istio JWT is **not** a production-ready feature, and API might change.

These are sample APIRule custom resources of both Ory Oathkeeper and Istio JWT access strategy configuration for a service.

<div tabs name="api-rule">
  <details>
  <summary>
  Ory Oathkeeper
  </summary>

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  host: foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - $ISSUER1
              - $ISSUER2
            jwks_urls:
              - $JWKS_URI1
              - $JWKS_URI2
            required_scope:
              - "test"
            target_audience:
              - "example.com"
              - "example.org"
            token_from:
              header: X-JWT-Assertion
```

  </details>
  <details>
  <summary>
  Istio
  </summary>

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
  namespace: $NAMESPACE
spec:
  gateway: kyma-system/kyma-gateway
  host: foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: $ISSUER1
              jwksUri: $JWKS_URI1
              fromHeaders:
              - name: X-JWT-Assertion
                prefix: "Kyma "
            - issuer: $ISSUER2
              jwksUri: $JWKS_URI2
              fromParameters:
              - "jwt_token"
            authorizations:
            - requiredScopes: ["test"]
              audiences: ["example.com", "example.org"]
            - requiredScopes: ["read", "write"]
```

  </details>
</div>

>**CAUTION:** Both `jwks_urls` and `trusted_issuers` must be valid `https` URLs.

>**CAUTION:** You can define multiple JWT issuers, but each of them must be unique.

>**CAUTION:** We support only a single `fromHeader` **or** a single `fromParameter` for a JWT issuer.

## Configuration of properties handling in Ory Oathkeeper and Istio resources

When you use Ory Oathkeeper, APIRule JWT access strategy configuration is translated directly as [authenticator configuration](https://www.ory.sh/docs/oathkeeper/api-access-rules#handler-configuration) in the [Ory Oathkeeper access rule CR](https://www.ory.sh/docs/oathkeeper/api-access-rules). See the official Ory Oathkeeper [JWT authenticator documentation](https://www.ory.sh/docs/oathkeeper/pipeline/authn#jwt) to learn more.

With Istio JWT access strategy, for each `authentications` entry, an Istio's [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) resource is created, and for each `authorizations` entry, an [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) resource is created.

## Corresponding JWT configuration properties in Ory Oathkeeper and Istio

This table lists all possible configuration properties of the Ory Oathkeeper JWT access strategy and their corresponding properties in Istio:

| Ory Oathkeeper | Required | | Istio | Required |
|-|:-:|-|-|:-:|
| **jwks_urls** | **YES** | &rarr; | **authentications.jwksUri** | **YES** |
| **trusted_issuers** | **NO** | &rarr; | **authentications.issuer** | **YES** |
| **token_from.header** | **NO** | &rarr; | **authentications.fromHeaders.name**<br/>**authentications.fromHeaders.prefix** | **NO** |
| **token_from.query_parameter** | **NO** | &rarr; | **authentications.fromParams** | **NO** |
| **token_from.cookie** | **NO** | &rarr; | *Not Supported* | **-** |
| **target_audience** | **NO** | &rarr; | **authorizations.audiences** | **NO** |
| **required_scope** | **NO** | &rarr; | **authorizations.requiredScopes** | **NO** |
| **scope_strategy** | **NO** | &rarr; | *Not Supported* | **-** |
| **jwks_max_wait** | **NO** | &rarr; | *Not Supported* | **-** |
| **jwks_ttl** | **NO** | &rarr; | *Not Supported* | **-** |
| **allowed_algorithms** | **NO** | &rarr; | *Not Supported* | **-** |

For more details, check the description of Istio JWT configuration properties in the [APIRule CR documentation](https://github.com/kyma-project/api-gateway/blob/main/docs/api-rule-cr.md#istio-jwt-configuration).

## How Istio differs from Ory Oathkeeper JWT access strategy

Istio JWT access strategy only supports `header` and `cookie` mutators. For more information, take a look at the [APIRule CR reference documentation](https://github.com/kyma-project/api-gateway/blob/main/docs/api-rule-cr.md#mutators).

Istio doesn't support regex type of path matching in Authorization Policies, which are supported by Ory Oathkeeper rules and by Virtual Service.

Istio doesn't support configuring a JWT token from `cookie`, and Ory Oathkeeper does. Istio supports only `fromHeaders` and `fromParams` configurations.

Using Istio as JWT access strategy requires the workload behind the service to be in the service mesh, for example, to have the Istio proxy. To learn how to add workloads to the Istio service mesh read the [Istio documentation](https://istio.io/latest/docs/ops/common-problems/injection/).

## Change of status `401` to `403` when calling an endpoint without `Authorization` header

With the introduction of Istio JWT the calls without a token will now fail on `AuthorizationPolicy` that forces `requestPrincipals` to not be empty. This results in a behaviour of requests without an `Authorization` header to fail on `AuthorizationPolicy` validation step, instead of `RequestAuthentication` that allows tokens without the header. In conclusion this behaviour will make the request fail with code `403` instead of `401`, which was the case for Oathkeeper JWT. 

## Blocking of in cluster connectivity to an endpoint

With the introduction of Istio JWT the validation of requests now happens in context of the target Pod (i.e. istio-sidecar). This behaviour results in in-cluster requests that where allowed in context of `ory-oathkeeper` (as request validation happened in context of `ory-oathkeeper` Pod) to fail if not explicitly allowed.


