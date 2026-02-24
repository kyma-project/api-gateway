# Configure the **jwt** Access Strategy

This page shows how to secure an APIRule with JSON Web Tokens (JWTs) using the **jwt** access strategy. The goal is to keep configuration simple while still giving you control over who can call your workload.

You can use this access strategy only with the Istio JWT configuration and you can define only one issuer per APIRule rule.

---

## 1. How It Works

When you use the `jwt` access strategy on a rule:

- API Gateway checks that the request contains a JWT.
- Istio validates the token:
  - It checks the issuer (`iss` claim).
  - It verifies the signature using the JWKS from the configured URL.
- Optionally, API Gateway checks additional requirements, such as scopes and audiences.

If any of these checks fail, the request is rejected before it reaches your Service.

---

## 2. Quick Start

This is the smallest configuration that protects a path with JWT. Any request that contains a **valid** JWT from the configured issuer is allowed.

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

In this configuration, API Gateway expects a JWT issued by `https://example.com`, and Istio uses the `jwksUri` to fetch the public keys and verify the token signature. Because you did not define any additional authorization rules, every valid token from that issuer is treated as authorized.

---

## 3. Configure Where the Token Comes From

By default, Istio reads the token from the `Authorization: Bearer <token>` header. You can override this behavior and read tokens from custom headers or query parameters using the `authentications` section.

```yaml
jwt:
  authentications:
    - issuer: https://example.com
      jwksUri: https://example.com/.well-known/jwks.json
      fromHeaders:
        - name: X-JWT-Assertion
          prefix: "Kyma "
      fromParameters:
        - "access_token"
```

Fields:
- `issuer` – The expected value of the JWT `iss` claim.
- `jwksUri` – The URL where Istio fetches the JWKS (public keys) to verify the token signature.
- `fromHeaders` – Optional. One or more HTTP headers to read the token from. Use `prefix` to remove a static prefix before parsing the token.
- `fromParameters` – Optional. One or more URL query parameter names to read the token from.

Under the hood, each entry in `authentications` creates a corresponding `requestPrincipals` value in Istio’s `AuthorizationPolicy`. Each value has the form `<ISSUER>/*`.

> [!NOTE]
> You can configure at most one issuer per APIRule rule when using the `jwt` access strategy.

---

## 4. Configure What the Token Must Contain

If you only configure `authentications`, any valid token from the issuer is allowed. To restrict access further, use `authorizations` to check scopes and audiences.

```yaml
jwt:
  authorizations:
    - requiredScopes: ["read", "write"]
      audiences: ["app1"]
```

Fields:
- `requiredScopes` – Optional. The token must contain **all** listed scopes in one of the `scp`, `scope`, or `scopes` claims.
- `audiences` – Optional. The token’s `aud` claim must contain **all** listed audiences.

Behavior:
- If `authorizations` is **not** defined, a request is authorized as long as the JWT is valid for the configured issuer.
- If multiple authorization entries are defined, the request is allowed if **at least one** entry matches.

---

## 5. End-to-End Flow

1. The client sends a request with a JWT (in a header or query parameter).
2. Istio validates the token using the `issuer` and `jwksUri` from `authentications`.
3. If validation succeeds, Istio sets the `requestPrincipal` to `<ISSUER>/*`.
4. API Gateway’s generated `AuthorizationPolicy` evaluates the `authorizations` rules:
   - If no `authorizations` are defined, any valid token from the issuer is enough.
   - If `authorizations` are defined, the token must satisfy at least one combination of `requiredScopes` and `audiences`.

Once these checks pass, the request is forwarded to your Service.


### Authentications
Under the hood, an authentications array creates a corresponding **requestPrincipals** array in the Istio’s Authorization Policy resource. Every **requestPrincipals** string is formatted as `<ISSUSER>/*`.

### Authorizations
The authorizations field is optional. When not defined, the authorization is satisfied if the JWT is valid. You can define multiple authorizations for an access strategy. The request is allowed if at least one of them is satisfied.

The **requiredScopes** and **audiences** fields are optional. If the **requiredScopes** field is defined, the JWT must contain all the scopes in the scp, scope, or scopes claims to be authorized. If the **audiences** field is defined, the JWT has to contain all the audiences in the aud claim to be authorized.