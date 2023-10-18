# API Rule dependencies

## Istio

APIRules require Istio installed on the cluster. This is required as the APIRule controller creates the custom resources `VirtualService`, `AuthorizationPolicy` and `RequestAuthentication` provided by Istio.

## Ory Oathkeeper

> NOTE: Ory Oathkeeper is deprecated. This part is subject to changes in the future.

APIRule controller will create a `Rule` Custom Resource when an `APIRule` with access strategy other than `allow` is used. As so `Ory Oathkeeper` with `Ory Oathkeeper Maester` is required for the controller.

Oathkeeper can be installed by installing `API Gateway` module.