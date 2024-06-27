# Migration of APIRule v1beta1 to v2alpha1

This document describes the outcome of the PoC to verify a migration of APIRule from v1beta1 to v2alpha1.

## Process from user point of view
- Enable Istio injection in the namespace and deploy a workload, e.g. httpbin
```bash
  kubectl label namespace default istio-injection=enabled --overwrite
  kubectl create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
```

- Create an APIRule v1beta1 securing httpbin with JWT. At this moment Oathkeeper is handling the JWT validation.
```bash
kubectl apply -f - <<EOF
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: httpbin-rule
  namespace: default
  labels:
    app: httpbin
spec:
  gateway: kyma-system/kyma-gateway
  host: httpbin
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /.*
      methods: [ "GET", "POST"]
      mutators: [ ]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - https://kymagoattest.accounts400.ondemand.com
            jwks_urls:
              - https://kymagoattest.accounts400.ondemand.com/oauth2/certs
EOF
```

- Create an APIRule v2alpha1 securing httpbin with JWT. At this moment JWT validation will be migrated from Oathkeeper to Istio.
```bash
kubectl apply -f - <<EOF
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: httpbin-rule
  namespace: default
  labels:
    app: httpbin
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - httpbin
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /.*
      methods: [ "GET", "POST"]
      jwt:
        authentications:
          - issuer: "https://kymagoattest.accounts400.ondemand.com"
            jwksUri: "https://kymagoattest.accounts400.ondemand.com/oauth2/certs"
EOF
```

## Reconciliation process of the APIRule
1. APIRule v1beta1 with Oathkeeper JWT configuration is applied
  - APIRule is reconciled and existing Ory reconciliation processing is used to create resources.
2. APIRule v2alpha1 with Istio JWT configuration is applied
  - In this case the conversion webhook adds the annotation `gateway.kyma-project.io/original-version: v2alpha1` to the APIRule.
  - If the controller detects the `gateway.kyma-project.io/original-version: v2alpha1` annotation on an APIRule it will execute a migration reconciliation.
  - The migration reconciliation will create the Virtual Service pointing to Ory, the Istio Authorization Policy and Request Authentication resources and 
     the Ory Access Rule. For the Access Rule a migration needs to be implemented since the APIRule v2alpha1 does not use the Ory configuration structure and naming.
  - If no error occurred and those resources were applied successfully, the controller will add a migration marker annotation `gateway.kyma-project.io/migration: v2alpha1` to the APIRule 
    if it does not exist and requeue the APIRule for the next reconciliation after 1 minute. We use the 1-minute time, so there is enough time for Istio to propagate the configuration.

3. APIRule v2alpha1 with Istio JWT configuration is reconciled after 1 minute
  - If the controller detects the `gateway.kyma-project.io/original-version: v2alpha1` annotation on an APIRule it will execute a migration reconciliation.
  - If the migration reconciliation detects the migration marker annotation `gateway.kyma-project.io/migration: v2alpha1` on the APIRule it will apply the Virtual Service pointing to the service directly instead of Oathkeeper, 
    the Istio Authorization Policy and Request Authentication resources and still the Ory Access Rule. The Ory Access Rule should be removed in the future, but there is a risk that there is a short time when the VS is still pointing to 
    Oathkeeper, but the AccessRules are already deleted. Therefore, it's more secure to keep the AccessRule for now and remove it with Ory-related code.

4. Removal of Ory Oathkeeper
  - Remove the Ory reconciliation processing and migration reconciliation processing from the controller.
  - Use AccessRuleProcessor in istio package to handle the deletion.
  - Remove migration marker annotation `gateway.kyma-project.io/migration: v2alpha1` from the APIRule.

### Implementation details
In the reconciler there are Istio and Ory reconciliation processing implemented. They can be switched using the If the `api-gateway-config` ConfigMap.
The reconciliation processing uses different processors to handle the creation, update and deletion of different resources, VirtualService, AccessRules, etc.
To manage the migration, a new reconciliation processing can be implemented that uses existing processors from the `ory` and `istio` package.
Additionally, a reconciliation processing needs validators for access strategies and JWT, but we might be able to drop some of the existing implementation, since a lot of
validation for APIRule v2alpha1 is already done in the CRD. 
At the point when the migration processing is implemented, the Istio processing is only used when it is explicitly enabled in the `api-gateway-config` ConfigMap.

## Verification
The steps from the user scenario were followed and additionally, after applying the v1beta1 APIRule,
The approach from the [Istio Zero downtime upgrade test](https://github.com/kyma-project/istio/issues/429) with an additional `-H "Authorization: Bearer $ACCESS_TOKEN"` was used
to send many parallel requests to httpbin to verify that no downtime occurs during the migration. The migration was successful and no downtime was observed.
