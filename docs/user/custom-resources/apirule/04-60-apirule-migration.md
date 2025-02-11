# APIRule Migration Procedure for Switching from `v1beta1` to `v2`

Authorization and authentication mechanism implemented in `v2` APIRule uses Istio RequestAuthentication and Istio AuthorizationPolicy instead of Ory Oathkeeper. 
As a result, temporal downtime could occur as there is a propagation delay for the Istio configuration to apply to the Envoy sidecar proxies.
To make sure that the migration can be completed without any downtime, a migration procedure has been implemented as part of APIRule reconciliation.

## Possible Scenarios

- In case you use Istio JWT as the authentication mechanism in version `v1beta1`, no special steps are required for the migration. Updating the version of an existing APIRule to `v2` will not change the authentication mechanism, as it was already using Istio JWT.
- In case you use Ory Oathkeeper as the authentication mechanism in version `v1beta1`, updating the APIRule to `v2` initiates the migration procedure described below.

## Migration Procedure

The migration procedure consists of the following steps, which are executed in a time-separated manner, with a one-minute delay between each step:
1. The resource owner updates the APIRule to version `v2`. As an immediate result, new Istio Authorization Policy and Istio Authentication Policy resources are created.
2. The Istio VirtualService resource is updated to point directly to target Service, bypassing Ory Oathkeeper.
3. The Ory Oathkeeper resource is deleted.
