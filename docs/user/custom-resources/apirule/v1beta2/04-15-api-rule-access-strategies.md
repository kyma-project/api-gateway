# APIRule v1beta2 Access Strategies

APIRule allows you to define the security configuration for an exposed endpoint using the concept of access strategies. You can specify access strategies in the **rules.accessStrategies** section of an APIRule.

Every **accessStrategy** contains two fields: **handler** and **config**. These fields determine which handler should be used and provide configuration options specific to the selected handler. The supported handlers for APIRule `v1beta2` are `no_auth` and `jwt`.

## Handler Configuration

### The `no_auth` Handler

The intended functionality of this handler is to provide a simple configuration for exposing workloads. It does not use Oathkeeper configuration and instead relies only on Istio VirtualService.

The `no_auth` handler only allows access to the specified HTTP methods of the exposed workload. You must not configure the **config** field when using this handler.

```yaml
rules:
  - path: /headers
    methods: ["GET"]
    noAuth: true
```

### The `jwt` Handler

By default, the `jwt` handler is configured in the same way as in the [Ory Oathkeeper JWT authenticator configuration](https://www.ory.sh/docs/oathkeeper/pipeline/authn#jwt). However, you can also use this handler with the Istio JWT configuration currently being developed. To learn more about this functionality, see [JWT Access Strategy](04-20-apirule-istio-jwt-access-strategy.md).

```yaml
rules:
  - path: /headers
    methods: ["GET"]
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
      authorizations:
        - audiences: ["app1"]
```
