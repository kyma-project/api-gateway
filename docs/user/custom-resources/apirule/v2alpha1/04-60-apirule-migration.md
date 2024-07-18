# APIRule migration procedure for switching from v1beta1 to v2alpha1

## Overview

With the introduction of `v2alpha1` APIRule the authorization and authentication mechanism will switch from using `Ory Oathkeeper` to `Istio AuthorizationPolicy` and `Istio RequestAuthentication`.
As a result of this, temporal downtime could happen as there is a propagation delay for the Istio configuration to apply to the Envoy sidecar proxies.
To make sure that the migration is done without any downtime, a migration procedure has been implemented as part of APIRule reconciliation.

## Possible scenarios

- In case Istio JWT was used previously as the authentication mechanism, no special steps are required for the migration. Updating the version of the existing APIRule to `v2alpha1` will not change the authentication mechanism, as it was already using Istio JWT.
- In case the APIRule was using Ory Oathkeeper as its authentication mechanism, the migration procedure described bellow will be triggered when updating the APIRule to `v2alpha1`.

## Migration procedure

The migration procedure happens in the following steps, which are executed in a time separated manner, with 1-minute delay between each step:
1. The APIRule is updated to `v2alpha1` by the resource owner. As an immediate result, new Istio Authorization Policy and Istio Authentication Policy resources are created.
2. The Istio VirtualService resource is updated to point directly to target Service, bypassing Ory Oathkeeper.
3. The Ory Oathkeeper resource is deleted.
