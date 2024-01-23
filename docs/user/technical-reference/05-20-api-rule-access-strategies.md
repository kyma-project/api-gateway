# API Rule access strategies

APIRule allows to configure how should an exposed endpoint be secured with the concept of `access strategies`. Those can be specified in the `rules` section of the APIRule, under `rules.accessStrategies`. 

Every `accessStrategy` contains `handler` field and `config`, which configure what handler should be used and configuration specific to the selected handler. The supported handlers are:
- allow
- noop
- unauthorized
- anonymous
- cookie_session
- bearer_token
- oauth2_client_credentials
- oauth2_introspection
- jwt

## Handler configuration

### `allow` handler

The intended functionality of this handler is a simple configuration allowing for easy workload exposure. This handler does not use any Oathkeeper configuration, depending only on Istio `VirtualService`.

The `allow` handler exposes access to the workload with all HTTP methods. With this handler `config` field must not be configured.

### `jwt` handler

This handler by default can be configured the same as in [Ory Oathkeeper JWT authenticator configuration](https://www.ory.sh/docs/oathkeeper/pipeline/authn#jwt). However, there is a possibility to use this handler with the currently in development `Istio JWT` configuration. If you are interested in that functionallity please have a look at [this document](../custom-resources/apirule/04-20-apirule-istio-jwt-access-strategy.md).

### Other handlers

Apart from `allow` handler and `jwt` with default configuration, all the other handlers are based on configuration documented in [Ory Oathkeeper authenticators](https://www.ory.sh/docs/oathkeeper/pipeline/authn), as Ory Oathkeeper is the component responsible for handling requests that use those handlers. As so, the configuration and capabilities of those handlers is the same as described in that document.

When using those handlers keep in mind that Ory stack as part of API Gateway is deprecated and will not be supported in the future.
