# Configure the **jwt** Access Strategy

Use the **jwt** access strategy to protect your workload with JSON Web Tokens (JWTs). With this strategy, API Gateway and Istio validate incoming tokens and only allow requests that satisfy your JWT configuration.

You can use this access strategy only with the Istio JWT configuration and define only one issuer per APIRule rule.

## Minimal Configuration

The minimal configuration secures a path with a single issuer and JWKS URI. Any request with a valid JWT from that issuer is allowed.

```yaml
...
rules:
  - path: /headers
    methods: ["GET"]
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
```

In this example:
- API Gateway expects a JWT from `https://example.com`.
- Istio fetches public keys from the `jwksUri` to verify the token signature.
- There is no additional authorization logic, so any valid token from that issuer is accepted.

## Configure Authentications

Use the `authentications` section to define **who** issues tokens and **where** they are read from.

```yaml
jwt:
  authentications:
    - issuer: https://example.com
      jwksUri: https://example.com/.well-known/jwks.json
      fromHeaders:
        - name: Authorization
          prefix: "Bearer "
      fromParameters:
        - "access_token"
```

- `issuer` – Expected `iss` claim in the JWT.
- `jwksUri` – URL of the JWKS document with public keys for signature verification.
- `fromHeaders` – Optional. One or more headers to extract the token from. Use `prefix` to strip values such as `Bearer ` or `Kyma `.
- `fromParameters` – Optional. One or more query parameter names to extract the token from.

Under the hood, the `authentications` array creates a corresponding `requestPrincipals` array in Istio’s `AuthorizationPolicy`. Each entry is formatted as `<ISSUER>/*`.

> [!NOTE]
> You can define only one issuer per APIRule rule when using the `jwt` access strategy.

## Configure Authorizations

Use the `authorizations` section to define **which** tokens are allowed, based on scopes and audiences.

```yaml
jwt:
  authorizations:
    - requiredScopes: ["read", "write"]
      audiences: ["app1"]
```

- `requiredScopes` – Optional. The JWT must contain **all** listed scopes in one of the `scp`, `scope`, or `scopes` claims.
- `audiences` – Optional. The JWT must contain **all** listed audiences in the `aud` claim.

Behavior:
- If `authorizations` is **not** defined, a request is authorized as long as the JWT is valid for the configured issuer.
- If multiple authorization entries are defined, the request is allowed if **at least one** entry matches.