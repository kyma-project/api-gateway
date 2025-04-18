# APIRule Migration Procedure for Switching from `v1beta1` to `v2`

Authorization and authentication mechanism implemented in `v2` APIRule uses Istio RequestAuthentication and Istio AuthorizationPolicy instead of Ory Oathkeeper. 
As a result, temporal downtime could occur as there is a propagation delay for the Istio configuration to apply to the Envoy sidecar proxies.
To make sure that the migration can be completed without any downtime, a migration procedure has been implemented as part of APIRule reconciliation.

Before any modifications, consult the documentation of changes introduced in the new version of APIRule `v2alpha1` in the [APIRule v2 Changes](04-70-changes-in-apirule-v2.md) document.

## Possible Scenarios

- In case you use Istio JWT as the authentication mechanism in version `v1beta1`, no special steps are required for the migration. Updating the version of an existing APIRule to `v2` will not change the authentication mechanism, as it was already using Istio JWT.
- In case you use Ory Oathkeeper as the authentication mechanism in version `v1beta1`, updating the APIRule to `v2` initiates the migration procedure described below.

## Migration Procedure

The migration procedure consists of the following steps, which are executed in a time-separated manner, with a one-minute delay between each step:
1.  As the resource owner, you must update the APIRule to version `v2`. As an immediate result, new Istio Authorization Policy and Istio Authentication Policy resources are created.
2. To retain the APIRule `v1beta1` CORS configuration, update the APIRule with the CORS configuration. For preflight requests to work, your APIRule needs to explicitly allow the `"OPTIONS"` method in the **rules.methods** field.
3. To retain APIRule `v1beta1` internal traffic policy, apply the following AuthorizationPolicy. Remember to change the selector label to the one pointing to the target workload:
    ```yaml
    apiVersion: security.istio.io/v1
    kind: AuthorizationPolicy
    metadata:
      name: allow-internal
      namespace: ${NAMESPACE}
    spec:
      selector:
        matchLabels:
          ${KEY}: ${TARGET_WORKLOAD}
      action: ALLOW
      rules:
      - from:
        - source:
            notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
    ```
4. The Istio VirtualService resource is updated to point directly to the target Service, bypassing Ory Oathkeeper.
5. The Ory Oathkeeper resource is deleted.
