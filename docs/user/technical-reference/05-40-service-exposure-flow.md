## Request flow

This diagram illustrates the request flow for three cases:
  - Accessing secured resources with an OAuth2 token
  - Accessing secured resources with a JWT token
  - Accessing unsecured resources without a token

![request-flow](../../assets/api-gateway-request-flow.svg)

**TIP:** Learn how to [Configure authorizations](../custom-resources/apirule/04-50-apirule-authorizations.md). 

### Accessing secured resources with an OAuth2 token

The developer sends a request to access a secured resource with an OAuth2 access token issued for a registered client. The request is proxied by the Oathkeeper proxy. The proxy identifies the token as an OAuth2 access token and sends it to the registered Token Introspection endpoint in the Hydra OAuth2 server. The OAuth2 server validates the token and returns the outcome of the validation to Oathkeeper. If the validation is successful, Oathkeeper checks the token against the Access Rules that exist for the resource and authorizes the request. Upon successful authorization, the request is forwarded to the resource.

### Accessing secured resources with a JWT token

The developer sends a request to access a secured resource with a JWT token. The request is proxied by the Oathkeeper proxy. The proxy identifies the token as a JWT token and fetches the public keys required for token validation from the registered OpenID Connect-compliant identity provider. Oathkeeper uses these keys to validate the token. If the validation is successful, Oathkeeper checks the token against the Access Rules that exist for the resource and authorizes the request. Upon successful authorization, the request is forwarded to the resource.

### Accessing unsecured resources without a token

The developer sends a request to access a resource without a token. The request is proxied by the Oathkeeper proxy. The proxy checks if there are Access Rules created for the resource, and verifies if it can be accessed without a token. If the resource can be accessed without a token, the request is forwarded to the resource.