# JWT Validation
Short Description: Learn how to configure the **jwt** access strategy in an APIRule custom resource to secure your workload with JSON Web Tokens (JWTs). You can use this access strategy only with the Istio JWT configuration and define only one issuer per APIRule rule.

## Minimal Configuration
Minimal **jwt** configuration secures a path with a single issuer and JWKS URI.

In the following example, the API Gateway module only accepts JWTs issued by `https://example.com`, and Istio uses the configured **jwksUri** to fetch the public keys and verify the token signature. Because no additional authorization rules are defined, any valid token from that issuer is treated as authorized.

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

## Configure Authentications
Configure authentications to further define where the token comes from.

By default, Istio reads the token from `Authorization: Bearer <token>`. You can override this and use custom headers or query parameters with the authentications section:

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

The authentications section contains the following fields:
- **issuer** - The expected value of the JWT iss claim.
- **jwksUri** - The URL where Istio fetches the JWKS (public keys) to verify the token signature.
- (Optional) **fromHeaders** - One or more HTTP headers to read the token from. Use prefix to remove a static prefix before parsing the token.
- (Optional) **fromParameters** - One or more URL query parameter names to read the token from.

Under the hood, the authentications array creates a corresponding requestPrincipals array in Istio’s AuthorizationPolicy. Each entry is formatted as <ISSUER>/*.

## Configure Authorizations
Configure authentications to further define the token's scopes and audiences. If authorizations are not defined, a request is authorized as long as the JWT is valid for the configured issuer. If multiple authorization entries are defined, the request is allowed if at least one entry matches.

```yaml
jwt:
  authorizations:
    - requiredScopes: ["read", "write"]
      audiences: ["app1"]
```

The authorizations section contains the following fields:
- (Optional) **requiredScopes** - The token must contain all listed scopes in one of the **scp**, **scope**, or **scopes** claims.
- (Optional) **audiences** - The token’s **aud** claim must contain all listed audiences.