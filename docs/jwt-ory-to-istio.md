---
title: JWT access strategies with Ory Oathkeeper to Istio
---

For JWT access strategy we are transitioning from Ory Oathkeeper to Istio. Here is a description on how Ory Oathkeeper JWT access strategy configuration looks like and respectively differs to Istio.

## Comparison of Ory Oathkeeper and Istio JWT access strategy configurations

>**CAUTION:** Istio JWT is **not** a production-ready feature, and API might change.

These are sample APIRule custom resources of both Ory Oathkeeper and Istio JWT access strategy configuration for a service.

<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="Ory Oathkeeper">
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
```

  </details>
  <details>
  <summary label="Istio">
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
            - issuer: $ISSUER
              jwksUri: $JWKS_URI
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

## Configuration properties handling into Ory Oathkeeper and Istio resources

For Ory Oathkeeper, APIRule JWT access strategy configuration is translated directly as [authenticator configuration](https://www.ory.sh/docs/oathkeeper/api-access-rules#handler-configuration) in the [Ory Oathkeeper access rule CR](https://www.ory.sh/docs/oathkeeper/api-access-rules). More details are available in the official Ory Oathkeeper [JWT authenticator documentation](https://www.ory.sh/docs/oathkeeper/pipeline/authn#jwt).

For Istio, for each `authentications` entry we create an Istio's [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) resource and for each `authorizations` entry we create an [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) resource respectively.

## Oathkeeper-to-Istio JWT corresponding configuration properties

This table lists all the possible configuration properties of the Ory Oathkeeper JWT access strategy and their corresponding properties in Istio JWT:

| Ory Oathkeeper | Required | | Istio | Required |
|-|:-:|-|-|:-:|
| **jwks_urls** | **YES** | &rarr; | **authentications.jwksUri** | **YES** |
| **trusted_issuers** | **NO** | &rarr; | **authentications.issuer** | **YES** |
| **scope_strategy** | **NO** | &rarr; | **authorizations.requiredScopes** | **NO** |
| **target_audience** | **NO** | &rarr; | **authorizations.audiences** | **NO** |
| **required_scope** | **NO** | &rarr; | **authorizations.requiredScopes** | **NO** |
| **jwks_max_wait** | **NO** | &rarr; | *Not Supported* | **-** |
| **jwks_ttl** | **NO** | &rarr; | *Not Supported* | **-** |
| **allowed_algorithms** | **NO** | &rarr; | *Not Supported* | **-** |
| **token_from** | **NO** | &rarr; | **authentications.fromHeaders.name**<br/>**authentications.fromHeaders.prefix**<br/>**authentications.fromParams**| **NO** |

For more details on Istio JWT configuration properties please check our [APIRule CR documentation](https://github.com/kyma-project/api-gateway/blob/main/docs/api-rule-cr.md#istio-jwt-configuration).

## Differences and deprecation coming with Istio JWT access strategy

Only `header` and `cookie` mutators are supported with Istio JWT access strategy. For more info please take a look at our [APIRule CR](https://github.com/kyma-project/api-gateway/blob/main/docs/api-rule-cr.md#mutators) reference documentation.

Istio doesn't support regex type of path matching on Authorization Policies, which were suported on Ory Oathkeeper rules and are supported by Virtual Service.

Istio doesn't support JWT token from `cookie` configuration, which was supported with Ory Oathkeeper. Istio supports only `fromHeaders` and `fromParams` configurations.
