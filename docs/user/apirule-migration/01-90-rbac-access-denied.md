# Migrating multiple APIRules targeting same workload from `v1beta1` to `v2`: Resolving RBAC Access Denied Errors

This tutorial explains a scenario where multiple APIRules `v1beta1` expose the same workload using different host names. During migration, you may want to keep all endpoints available. To ensure your service requests are handled as intended this tutorial describes how to not obtain `RBAC: Access Denied` errors that can occur when accessing a target workload using `v1beta1` APIRule when other APIRule is already migrated to version `v2` and also targets same workload.

## Context
You have multiple APIRules `v1beta1` exposing the same workload but using different host values. During migration of these APIRules to version `v2`, you want to ensure that all endpoints remain accessible without downtime.
After migrating one of those APIRules from `v1beta1` to `v2`, requests by APIRules `v1beta1` endpoints will result in `HTTP/2 403 RBAC: Access Denied` errors when accessing the target workload. But requests to the migrated APIRule `v2` endpoint will work as expected.

Example of possible migration scenario where mixed versions APIRules target same workload. APIRule `example1` is already migrated to version `v2` while APIRule `example2` is still in version `v1beta1` and `403 RBAC: Access Denied` errors occur when accessing the target workload using APIRule `example2`.
```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: example1
spec:
  hosts: 
    - example1
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /post
      methods: ["POST"]
      noAuth: true
      service:
        name: httpbin
        namespace: default
        port: 8000
```

```bash
curl -ik -X POST https://example1/post
HTTP/2 200
```

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: example2
spec:
    host: example2
    service:
      name: httpbin
      namespace: test
      port: 8000
    gateway: kyma-gateway.kyma-system
    rules:
      - path: /post
        service:
          name: httpbin
          namespace: test
          port: 8000
        methods:
          - POST
        accessStrategies:
          - handler: noop
```

```bash
curl -ik -X POST https://example2/post
HTTP/2 403
RBAC: access denied%
```

## Cause

After migration, one of your APIRules is now in version `v2`, which operates using Istio subresources. The other APIRules exposing the same workload remain in version `v1beta1` and do not use Istio subresources, so requests from `v1beta1` APIRules to this endpoint are no longer allowed. Currently, access to the target workload is permitted only for policies that match the rules specified in the APIRule Custom Resource for version `v2`.

## Procedure
In order to allow traffic to the target workload from other APIRules `v1beta1` during migration, you must create an additional **temporary** AuthorizationPolicy.

## Access Strategies: no_auth, allow

You need to create an ALLOW-type AuthorizationPolicy to permit traffic to the target workload for requests originating from the ingress gateway, specifically for hosts exposed by `v1beta1` APIRules you want to migrate.

### Temporary AuthorizationPolicy Template
| Option                             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
|------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **{NAMESPACE}**                    | The namespace to which the AuthorizationPolicy applies. This namespace must include the target workload for which you allow traffic. The selector matches workloads in the same namespace as the AuthorizationPolicy.                                                                                                                                                                                                                                                                                                               |
| **{HOSTNAME}**                     | List all host namesthat are exposed by the `v1beta1` APIRules being migrated. These hostnames represent the external endpoints through which clients access the workload. List all relevant hostnames to ensure the AuthorizationPolicy allows traffic for each endpoint during migration.                                                                                                                                                                                                                                                                                                                                                                                                                       |
| **{LABEL_KEY}**: **{LABEL_VALUE}** | To further restrict the scope of the AuthorizationPolicy, specify label selectors that match the target workload. Replace these placeholders with the actual key and value of the label. The label indicates a specific set of Pods to which a policy should be applied. The scope of the label search is restricted to the configuration namespace in which the AuthorizationPolicy is present. <br>For more information, see [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/). |

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
   name: allow-migration
   namespace: {NAMESPACE}
spec:
  action: ALLOW
  rules:
  - from:
    - source:
        principals:
        - "cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"
    to:
    - operation:
        notHosts:
        - {HOSTNAME1}
        - {HOSTNAME2}
  selector:
    matchLabels:
      {LABEL_KEY}: {LABEL_VALUE}
```

### Sample Scenario

In this scenario, there are two `v1beta1` APIRules configured to expose the same HTTPBin service, each with a different host value. This example demonstrates how to maintain uninterrupted access to service endpoints during migration by applying a temporary AuthorizationPolicy.

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: example1
spec:
  host: example1
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /.*
      service:
        name: httpbin
        namespace: default
        port: 8000
      methods: ["POST", "GET"]
      accessStrategies:
        - handler: no_auth
---
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: example2
spec:
  host: example2
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /.*
      service:
        name: httpbin
        namespace: default
        port: 8000
      methods: ["POST", "GET"]
      accessStrategies:
        - handler: no_auth
```

To ensure a seamless migration to `v2` without any downtime, creation of a **temporary** AuthorizationPolicy was needed. This policy allows traffic from the ingress gateway to the target workload for all relevant hosts exposed by the `v1beta1` APIRules which you should specify in `hosts` field of AuthorizationPolicy. Applying this policy before starting the migration process ensures uninterrupted access to service endpoints during the migration.

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-migration
spec:
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
            - "cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"
      to:
        - operation:
            hosts:
            - example1.local.kyma.dev
            - example2.local.kyma.dev
  selector:
    matchLabels:
      app: httpbin
```
> [!NOTE] This AuthorizationPolicy is temporary and must be deleted after migration of all APIRules targeting this specicifc workload to `v2` is done. We don't recommend mixed version of APIRules for one workload. We suggest migrating all APIRules targeting same workload at one go.
> 

This way both APIRules `v1beta1` will work during migration.
We can migrate APIRules to `v2` following the steps in [migration guidelines](../apirule-migration/README.md).

After migrating all APIRules targeting same workload to `v2` delete the temporary AuthorizationPolicy.
```bash
kubectl delete authorizationpolicy <AUTHORIZATION_POLICY_NAME> -n <NAMESPACE>
```




## Access Strategies: noop, jwt, oauth2_introspection
You need to create an ALLOW-type AuthorizationPolicy to permit traffic to the target workload for requests originating from the Oathkeeper service, specifically for hosts exposed by `v1beta1` APIRules you want to migrate.

### Temporary AuthorizationPolicy Template
| Option                             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
|------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **{NAMESPACE}**                    | The namespace to which the AuthorizationPolicy applies. This namespace must include the target workload for which you allow traffic. The selector matches workloads in the same namespace as the AuthorizationPolicy.                                                                                                                                                                                                                                                                                                               |
| **{LABEL_KEY}**: **{LABEL_VALUE}** | To further restrict the scope of the AuthorizationPolicy, specify label selectors that match the target workload. Replace these placeholders with the actual key and value of the label. The label indicates a specific set of Pods to which a policy should be applied. The scope of the label search is restricted to the configuration namespace in which the AuthorizationPolicy is present. <br>For more information, see [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/). |


```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-migration
  namespace: {NAMESPACE}
spec:
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
            - "cluster.local/ns/kyma-system/sa/oathkeeper-maester-account"
  selector:
    matchLabels:
      {LABEL_KEY}: {LABEL_VALUE}
```

### Sample Scenario
For purposes of scenario for access strategies **jwt** APIRules `v1beta1` we consider you have 2 APIRules `v1beta1` exposing same HTTPBin service but using different host values.
```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: example1
spec:
  gateway: kyma-system/kyma-gateway
  host: example1.local.kyma.dev
  service:
    name: httpbin
    namespace: default
    port: 8000
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - "https://kymagoattest.accounts400.ondemand.com"
            jwks_urls:
              - "https://kymagoattest.accounts400.ondemand.com/oauth2/certs"
---
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: example2
spec:
  gateway: kyma-system/kyma-gateway
  host: example2.local.kyma.dev
  service:
    name: httpbin
    namespace: default
    port: 8000
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - "https://kymagoattest.accounts400.ondemand.com"
            jwks_urls:
              - "https://kymagoattest.accounts400.ondemand.com/oauth2/certs"
```
And you want to migrate to `v2` without any possible downtime so you create this temporary additional AuthorizationPolicy to allow traffic from the Oathkeeper service to the target workload.

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-migration
spec:
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
              - "cluster.local/ns/kyma-system/sa/oathkeeper-maester-account"
  selector:
    matchLabels:
      app: httpbin
```

> [!NOTE] This AuthorizationPolicy is temporary and must be deleted after migration of all APIRules targeting this specicifc workload to v2 is done. We don't recommend mixed version of APIRules for one workload for longer period of time.
>

This way both APIRules v1beta1 will work during migration. Now you can migrate first APIRule to v2 according with migration guidelines.

After migrating all APIRules targeting same workload to v2 delete the temporary AuthorizationPolicy.