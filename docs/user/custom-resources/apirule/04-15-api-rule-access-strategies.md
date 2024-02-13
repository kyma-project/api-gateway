# APIRule Access Strategies

APIRule allows you to define the security configuration for an exposed endpoint using the concept of access strategies. You can specify access strategies in the **rules.accessStrategies** section of an APIRule.

Every **accessStrategy** contains two fields: **handler** and **config**. These fields determine which handler should be used and provide configuration options specific to the selected handler. The supported handlers are:
- `allow`
- `no_auth`
- `noop`
- `unauthorized`
- `anonymous`
- `cookie_session`
- `bearer_token`
- `oauth2_client_credentials`
- `oauth2_introspection`
- `jwt`

## Handler Configuration

### The `allow` Handler

The intended functionality of this handler is to provide a simple configuration for exposing workloads. It does not use Oathkeeper configuration and instead relies only on Istio VirtualService.

The `allow` handler allows access to the exposed workload with all HTTP methods. You must not configure the **config** field when using this handler.

### The `no_auth` Handler

The intended functionality of this handler is to provide a simple configuration for exposing workloads. It does not use Oathkeeper configuration and instead relies only on Istio VirtualService.

The `no_auth` handler only allows access to the specified HTTP methods of the exposed workload. You must not configure the **config** field when using this handler.

### The `jwt` Handler

By default, the `jwt` handler is configured in the same way as in the [Ory Oathkeeper JWT authenticator configuration](https://www.ory.sh/docs/oathkeeper/pipeline/authn#jwt). However, you can also use this handler with the Istio JWT configuration currently being developed. To learn more about this functionality, see [JWT Access Strategy](04-20-apirule-istio-jwt-access-strategy.md).

### Other Handlers

Except for the `allow`, `no_auth`, and `jwt` handlers, which use the default configuration, all the other handlers are based on the configuration documented in [Ory Oathkeeper Authenticators](https://www.ory.sh/docs/oathkeeper/pipeline/authn). Ory Oathkeeper is responsible for handling requests that use these handlers so their configuration and capabilities align with what is described in the documentation.

When using those handlers keep in mind that Ory stack as part of API Gateway is deprecated and will not be supported in the future.
