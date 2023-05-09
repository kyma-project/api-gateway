---
title: JWT access strategies with Ory Oathkeeper to Istio
---

For JWT access strategy we are transitioning from Ory Oathkeeper to Istio. Here is a description on how Ory Oathkeeper JWT access strategy configuration looks like and respectively differs to Istio.

## Ory Oathkeeper JWT access strategy

This is a sample APIRule custom resource (CR) that specifies Ory Oathkeeper JWT access strategy for a service.

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

Configuration for JWT access strategy is translated directly as [authenticator configuration](https://www.ory.sh/docs/oathkeeper/api-access-rules#handler-configuration) in the [Ory Oathkeeper access rule CR](https://www.ory.sh/docs/oathkeeper/api-access-rules), for more details please refer to official Ory Oathkeeper [JWT authenticator documentation](https://www.ory.sh/docs/oathkeeper/pipeline/authn#jwt).

This table lists all the possible parameters of the Ory Oathkeeper JWT access strategy together with their descriptions:

| Field | Mandatory  | Description |
|-|:-:|-|
| **jwks_urls** |  **YES**   | The URLs where Ory Oathkeeper can retrieve JSON Web Keys from for validating the JSON Web Token. Usually something like https://my-keys.com/.well-known/jwks.json. The response of that endpoint must return a JSON Web Key Set (JWKS). |
| **jwks_max_wait** | **NO** | The maximum time for which the JWK fetcher should wait for the JWK request to complete. After the interval passes, the JWK fetcher will return expired or no JWK at all. If the initial JWK request finishes successfully, it will still refresh the cached JWKs. Defaults to "1s". |
| **jwks_ttl** | **NO** | The duration for which fetched JWKs should be cached internally. Defaults to "30s". |
| **scope_strategy** | **NO** | Sets the strategy to be used to validate/match the scope. Supports "hierarchic", "exact", "wildcard", "none". Defaults to "none". |
| **trusted_issuers** | **NO** | The JWT must contain a value for claim iss that matches exactly (case-sensitive) one of the values of trusted_issuers. If no values are configured, the issuer will be ignored. |
| **target_audience** | **NO** | The JWT must contain all values (exact, case-sensitive) in the claim aud. If no values are configured, the audience will be ignored. |
| **allowed_algorithms** | **NO** | The signing algorithms are allowed. Defaults to RS256. |
| **required_scope** | **NO** | The scope of the JWT. It will checks for claims scp, scope, scopes in the JWT when validating the scope as that claim isn't standardized. |
| **token_from** | **NO** | The location of the bearer token. If not configured, the token will be received from a default location - 'Authorization' header. One and only one location (`header`, `query_parameter`, or `cookie`) must be specified. |

## Istio JWT access strategy configuration

>**CAUTION:** Istio JWT is **not** a production-ready feature, and API might change.

This is a sample APIRule custom resource (CR) that specifies Istio JWT access strategy for a service.

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

This table lists all the possible parameters of the Istio JWT access strategy together with their descriptions:

| Field | Mandatory | Description |
|-|:-:|-|
| **authentications** | **YES** | List of authentication objects. |
| **authentications.issuer** | **YES** | Identifies the issuer that issued the JWT. <br/>Must be an URL starting with `https://`. |
| **authentications.jwksUri** | **YES** | URL of the provider’s public key set to validate the signature of the JWT. <br/>Must be an URL starting with `https://`. |
| **authentications.fromHeaders** | **NO** | List of headers from which the JWT token is taken. |
| **authentications.fromHeaders.name** | **YES** | Name of the header. |
| **authentications.fromHeaders.prefix** | **NO** | Prefix used before the JWT token. The default is "`Bearer `".|
| **authentications.fromParams** | **NO** | List of parameters from which the JWT token is taken. |
| **authorizations** | **NO** | List of authorization objects. |
| **authorizations.requiredScopes** | **NO** | List of required scope values for the JWT. |
| **authorizations.audiences** | **NO** | List of audiences required for the JWT. |

>**CAUTION:** You can define multiple JWT issuers, but each of them must be unique.

>**CAUTION:** Currently, we support only a single `fromHeader` **or** a single `fromParameter`. Specifying both of these fields for a JWT issuer is not supported.

## Differences and deprecation coming with Istio JWT access strategy

* Only `header` and `cookie` mutators are supported with Istio JWT access strategy. For more info please take a look at our [APIRule CR](https://github.com/kyma-project/api-gateway/blob/main/docs/api-rule-cr.md#mutators) reference documentation.

* Istio doesn't support regex type of path matching on Authorization Policies, which were suported on Ory Oathkeeper rules and are supported by Virtual Service.

* Istio doesn't support JWT token from `cookie` configuration, which was supported with Ory Oathkeeper. Istio supports only `fromHeaders` and `fromParams` configurations.
