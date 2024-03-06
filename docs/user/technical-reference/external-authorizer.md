# API proposal for configuration of External Authorizer based authorization in API Rules

## Considerations

### JWT claim based authorization

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

As so we should allow combination of this handler with others.

### Support for multiple external authorizers combinations

We need to consider whether a configuration that will use multiple external authorizers on one path is valuable. Technically, this is possible to do, as all CUSTOM policies will need to result in `allow` response for the request to be allowed.

## API Proposal

Considering the logic of CUSTOM Istio AuthorizationPolicy, suggestion would be to not handle external authorizer as an additional access strategy, but have a separate field (array) in `rule`, that would configure only external authorizer.

This could be achieved by adding a `spec.rules[*].externalAuthorizers` array, that would hold an array of strings, that are the names of the authorizers defined in Istio mesh configuration.

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
      externalAuthorizers:
      - "oauth2-proxy" # Assuming that we will support an array
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: https://example.com
              jwksUri: https://example.com/.well-known/jwks.json            
            authorizations:
            - audiences: ["app1"]
```

This would create two AuthorizationPolicies:

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
  selector:
    matchLabels:
      app: httpbin
```

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: ext-authz
spec:
  action: ALLOW
  rules:
  - to:
    - operation:
        paths:
        - /headers
    when:
      - key: request.auth.claims[aud]
        values:
          - app1
  selector:
    matchLabels:
      app: httpbin
```

And a RequestAuthentication:

```yaml
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: httpbin
spec:
  jwtRules:
  - issuer: https://example.com
    jwksUri: https://example.com/.well-known/jwks.json
  selector:
    matchLabels:
      app: httpbin
```

This ensures that the user accessing the resource authenticates against `oauth2-proxy`, and the resulting JWT has `app1` audience.

