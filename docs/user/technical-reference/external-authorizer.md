# API proposal for configuration of External Authorizer based authorization in API Rules

The decided name for the `accessStrategy` is `extAuth`. This follows the naming convention we have for previous strategies. The name will follow camel case convention, as `no_auth` will change name to `noAuth` in the future.

## Considerations

### Should we support combining `extAuth` with `jwt` access strategy?

Because the Authorization Policy that enables External Authorizer uses `action: CUSTOM`, there is a possibility to mix up External Authorizer handler with different handlers (especially with Istio based JWT). This is possible because `CUSTOM` actions are evaluated independently from others, as described in [Istio documentation](https://istio.io/latest/docs/reference/config/security/authorization-policy). This will allow the customer to have a setup that performs both authentication with a OAuth2 Authorization Code flow, as well as authorization based on the presented JWT.

Especially, the following example configuration is possible:

- An `AuthorizationPolicy` enabling External Authorizer:

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: ext-authz
spec:
  action: CUSTOM
  provider:
    name: oauth2-proxy
  rules:
  - to:
    - operation:
        paths:
        - /headers
```

- and an `AuthorizationPolicy` restricting the access on a claim based strategy:

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: require-claim
spec:
  action: ALLOW
  rules:
  - to:
      - operation:
        paths:
          - /headers
    when:
      - key: request.auth.claims[some_claim]
        values:
          - some_value
```

- and an additional `RequestAuthentication` that makes sure Istio recognizes the issuer:
```yaml
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: httpbin
spec:
  jwtRules:
  - issuer: https://example.com
    jwksUri: https://example.com/.well-known/jwks.json
```

**Decision**
We decided to enforce that the user **MUST** have **ONE**, and **ONLY ONE** `accessStrategy` per every entry in `spec.rules`. As so, we don't support mixing up `extAuth` and `jwt`.
Instead, we would like to allow configuration in scope of the `extAuth` strategy, and creating a `ALLOW` AuthorizationPolicy based on that configuration.
In result, the proposed API would look as follows:

```yaml
accessStrategy: # Validation: there needs to be one access strategy, and only one
  extAuth:
    name: oauth2-proxy
    restrictions: # Feedback about name and structure appreciated
      # Will most likely have the same structure as in `jwt` access strategy
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json            
      authorizations:
        - audiences: ["app1"]
```

### Should we support multiple external authorizers?

We need to consider whether a configuration that will use multiple external authorizers on one path is valuable. Technically, this is possible to do, as all CUSTOM policies will need to result in `allow` response for the request to be allowed.

**Decision**
We decided to not support multiple external authorizers. Supporting multiple external authorizers will introduce additional complexity in the API, that might lead to unexpected behaviour or/and confuse the user.
If there is a use case for configuring multiple external authorizers, there is still a posibillity of creating the required `AuthorizationPolicies` themselves.

## API Proposal

We have discussed the api in context of clearing up the API in `v1beta2/v1` API versions. We decided that `accessStrategy` will hold a single entry, either `extAuth`, `jwt` or `noAuth`. The users **MUST** define **ONE**, and **ONLY ONE** access strategy in every `rule` in `spec.rules`.

A sample using the proposed API would look as follows:

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  host: httpbin
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      mutators:
        - ***
      accessStrategy: # Validation: there needs to be one access strategy, and only one
        extAuth:
          name: oauth2-proxy # Validation: Check if there is that authorizer in Istio mesh config
          restrictions:
            authentications:
              - issuer: https://example.com
                jwksUri: https://example.com/.well-known/jwks.json            
            authorizations:
              - audiences: ["app1"]
        ## OR
        jwt:
          authentications:
            - issuer: https://example.com
              jwksUri: https://example.com/.well-known/jwks.json            
          authorizations:
            - audiences: ["app1"]
        ## OR
        noAuth: true # If you have better idea for structure of `noAuth` please comment :)
```
