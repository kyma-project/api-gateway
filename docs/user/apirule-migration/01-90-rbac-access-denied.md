# Migrating multiple APIRules targeting same workload from `v1beta1` to `v2`: Resolving RBAC Access Denied Errors

This tutorial explains a scenario where multiple APIRules `v1beta1` are specified to expose the same workload using different host names. During migration, you may want to keep all endpoints available this is why an additional temporary AuthorizationPolicy is required. To ensure your service requests are handled as intended this tutorial describes how to not obtain `RBAC: Access Denied` errors that can occur when accessing a target workload using `v1beta1` APIRule when other APIRule is already migrated to version `v2` and also targets same workload.

## Context
You have multiple APIRules `v1beta1` exposing the same workload but using different host values. During migration of these APIRules to version `v2`, you want to ensure that all endpoints remain accessible without downtime.

After migrating first of those APIRules from `v1beta1` to `v2`, requests by APIRules `v1beta1` endpoints will result in `HTTP/2 403 RBAC: Access Denied` errors when accessing the target workload. But requests to the migrated APIRule `v2` endpoint will work as expected.

`HTTP/2 403 RBAC: Access Denied` erros are caused by Istio subresources created by migrated APIRule `v2`. The other APIRules exposing the same workload remain in version `v1beta1` and do not use Istio subresources, so requests from `v1beta1` APIRules to this endpoint are no longer allowed. Currently, access to the target workload is permitted only for requests that match the rules specified in the APIRule Custom Resource for version `v2`.

For example, suppose you have applied the following two `v1beta1` APIRules configured to expose the same HTTPBin service, each with a different host value and you want to maintain uninterrupted access to service endpoints during migration.

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: example1
  namespace: test
spec:
  host: example1
  service:
    name: httpbin
    namespace: default
    port: 8000
  gateway: kyma-gateway.kyma-system
  rules:
    - path: /post
      methods: ["POST"]
      accessStrategies:
        - handler: no_auth
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
  namespace: test
spec:
  host: example2
  service:
    name: httpbin
    namespace: default
    port: 8000
  gateway: kyma-gateway.kyma-system
  rules:
    - path: /post
      methods: ["POST"]
      accessStrategies:
        - handler: no_auth
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
To ensure a seamless migration to `v2` without any downtime, creation of a **temporary** AuthorizationPolicy before applying first migrated APIRule `v2` is a must.

To learn how to do this, follow the procedure.

> [!NOTE] We don't recommend mixed version of APIRules for one workload. We suggest migrating all APIRules targeting same workload at one go to version `v2`. Creating this temporary AuthorizationPolicy will block internal traffic to the workload which migration to APIRule `v2` will also cause. Please consider this when planning your migration. Go through every step of migration or if you want to allow internal traffic to the workload before migration look at this [troubleshooting guide](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-80-blocked-in-cluster-communication), but beware that any ALLOW-type AuthorizationPolicy will automatically block traffic that doesn't met the requirements of policies specified there. We recommend this course of action to apply this temporary AuthorizationPolicy (block in-cluster communication, unblock v1beta1 exposure external traffic), migrate APIRules to `v2`, delete this temporary AuthorizationPolicy and then apply any other AuthorizationPolicy you need to allow internal traffic to the workload again.


## Procedure
1. List hosts exposed by `v1beta1` APIRules targeting the same workload that will be migrated. In this example, the hosts are `example1` and `example2`.
2. Identify the label key and value for the target workload by checking the selector of the service being exposed by those APIRules. In this case, the selector for the `httpbin` service in the `default` namespace is `app: httpbin`.

    You can use this command:
    ```bash 
    kubectl get service <SERVICE_NAME> -n <NAMESPACE> -o jsonpath='{.spec.selector}'
    ```
   
    In our scenario case:
    ```bash
   kubectl get service httpbin -n default -o jsonpath='{.spec.selector}'
    {"app":"httpbin"}%
    ```
3. Create a temporary AuthorizationPolicy using the gathered information. This policy will allow traffic from both the Ory Oathkeeper service and the Istio ingress gateway to the target workload for the specified hosts.

    | Option                             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
    |------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
    | **{NAMESPACE}**                    | The namespace to which the AuthorizationPolicy applies. This namespace must include the target workload for which you allow traffic. The selector matches workloads in the same namespace as the AuthorizationPolicy.                                                                                                                                                                                                                                                                                                                |
    | **{HOSTNAME}**                     | List all host names that are exposed by the `v1beta1` APIRules being migrated. These hostnames represent the external endpoints through which clients access the workload. List all relevant hostnames to ensure the AuthorizationPolicy allows traffic for each endpoint during migration.                                                                                                                                                                                                                                          |
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
        - from:
            - source:
                principals:
                  - "cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"
          to:
            - operation:
                hosts:
                  - {HOSTNAME1}
                  - {HOSTNAME2}
      selector:
        matchLabels:
          {LABEL_KEY}: {LABEL_VALUE}
    ```

    This policy allows traffic from the ingress gateway to the target workload for all specified hosts and all traffic from the Ory Oathkeeper service. Applying this policy before starting the migration process ensures uninterrupted access to service endpoints during the migration.

    In our example, the temporary applied AuthorizationPolicy would look like this:
    ```yaml
    apiVersion: security.istio.io/v1
    kind: AuthorizationPolicy
    metadata:
      name: allow-migration
      namespace: default
    spec:
      action: ALLOW
      rules:
        - from:
            - source:
                principals:
                  - "cluster.local/ns/kyma-system/sa/oathkeeper-maester-account"
        - from:
            - source:
                principals:
                  - "cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"
          to:
            - operation:
                hosts:
                  - example1
                  - example2
      selector:
        matchLabels:
          app: httpbin
    ```

4. Migrate APIRules `v1beta1` to `v2` following the steps in [migration guidelines](../apirule-migration/README.md). During this process, the temporary AuthorizationPolicy ensures that requests from both `v1beta1` and `v2` APIRules to the target workload are allowed.

> [!NOTE] This AuthorizationPolicy is temporary and must be deleted after migration of all APIRules targeting this specicifc workload to `v2` is done. We don't recommend mixed version of APIRules for one workload. We suggest migrating all APIRules targeting same workload at one go.
>

5. After migrating all APIRules targeting the same workload to `v2` is finished, delete the temporary AuthorizationPolicy.

    ```bash
    kubectl delete authorizationpolicy <AUTHORIZATION_POLICY_NAME> -n <NAMESPACE>
    ```
    
    In our example:
    ```bash
    kubectl delete authorizationpolicy allow-migration -n default
    ```
