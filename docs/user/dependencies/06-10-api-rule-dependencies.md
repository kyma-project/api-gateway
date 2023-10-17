# API Rule dependencies

## Istio

Creation of APIRules require the pre-requisite of Istio installed on the cluster as the controler may create `VirtualService`, `AuthorizationPolicy` and `RequestAuthentication`.

## Ory Oathkeeper

> NOTE: Ory Oathkeeper is deprecated. This part is subject to changes in the future.

APIRule controller will create a `Rule` Custom Resource when an `APIRule` with access strategy other than `allow` is used. As so `Ory Oathkeeper` with `Ory Oathkeeper Maester` is required for the controller.

Oathkeeper can be installed by installing `API Gateway` module.