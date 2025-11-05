# Migrating multiple APIRules targeting same workload from `v1beta1` to `v2`

This tutorial explains a scenario where multiple APIRules `v1beta1` are exposing the same workload using different host names. During migration, you may want to keep all endpoints available this is why an additional temporary AuthorizationPolicy is required. To ensure your service requests are handled as intended this tutorial describes how to create an additional, temporary AuthorizationPolicy to not obtain `RBAC: Access Denied` errors that can occur when accessing a target workload using `v1beta1` APIRule when other APIRule is already migrated to version `v2` and also targets the same workload.

## Context
You have multiple APIRules `v1beta1` exposing the same workload but using different host values. During migration of these APIRules to version `v2`, you want to ensure that all endpoints remain accessible without downtime.

After migrating one of those APIRules from `v1beta1` to `v2`, requests by APIRules `v1beta1` endpoints will result in `HTTP/2 403 RBAC: Access Denied` errors when accessing the target workload. But requests to the migrated APIRule `v2` endpoint will work as expected returning `HTTP/2 200 OK` responses.

`HTTP/2 403 RBAC: Access Denied` errors are caused by Istio subresources created by migrated APIRule `v2`. The other APIRules exposing the same workload remain in version `v1beta1` and do not use Istio subresources, so requests from `v1beta1` APIRules to this endpoint are no longer allowed. Currently, access to the target workload is permitted only for requests that match the rules specified in the APIRule Custom Resource for version `v2`.

For example, suppose a scenario you have applied the following two `v1beta1` APIRules configured to expose the same HTTPBin service, each with a different host value, and you want to maintain uninterrupted access to service endpoints during migration.

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
              - https://{IAS_TENANT}.accounts.ondemand.com
            jwks_urls:
              - https://{IAS_TENANT}.accounts.ondemand.com/oauth2/certs
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
              - https://{IAS_TENANT}.accounts.ondemand.com
            jwks_urls:
              - https://{IAS_TENANT}.accounts.ondemand.com/oauth2/certs
```
To ensure a seamless migration to `v2` without any downtime, creation of an **additional, temporary** AuthorizationPolicy before applying first migrated APIRule `v2` is a must.

To learn how to do this, follow the procedure.

> [!NOTE] We don't recommend mixed version of APIRules for one target workload. We suggest migrating all APIRules targeting same workload at one go to version `v2`. Creating this temporary AuthorizationPolicy will block internal traffic to the workload which migration to APIRule `v2` will anyway cause. Please consider this when planning your migration. Please get acquainted with every step of [migration tutorials](./README.md). Please beware that any ALLOW-type AuthorizationPolicy will automatically block traffic that doesn't meet the requirements of policies specified there. 
> 
> We recommend this course of action:  
> 1. Apply this temporary AuthorizationPolicy (will cause block in-cluster communication, unblock v1beta1 exposure external traffic during the migration). 
> 2. Migrate APIRules to `v2`. Accoridng to documentation [migration guidelines](../apirule-migration/README.md).
> 3. Delete this temporary AuthorizationPolicy.


## Procedure
1. Go through APIRules `v1beta1` which target same workload. List hosts from this APIRules which contains at least one of handlers `allow` or `no_auth`.  Specify those hosts in the FQDN format.

    > [!NOTE] If your APIRules `v1beta1` don't use `allow` or `no_auth` handlers you should skip this first point. Specifying those hosts is required to allow traffic from Istio ingress gateway to the target workload during migration only for APIRules with `allow` or `no_auth` access strategy. Other handlers like `jwt`, `oauth2_introspection` and `noop` use the Ory Oathkeeper service and traffic doesn't come directly from ingress gateway to the target workload.
    >

    In our example, we have two APIRules `v1beta1` with `no_auth` handler exposing the same `httpbin` service with different hosts: 
    - `example1`  
    - `example2`
    
    To obtain the FQDN format of the hosts, you need to obtain a default domain to do that you can use the following command:

   ```bash
   kubectl get gateway <GATEWAY_NAME> -n <GATEWAY_NAMESPACE> -o jsonpath='{.spec.servers[0].hosts}'
    ```
   
   In tutorial scenario case we have value in APIRules of gateway `kyma-gateway` in namespace `kyma-system`, so the command will look like this:

    ```bash
    kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'
   ["*.local.kyma.dev"]%
    ```
    Assuming the default domain is `local.kyma.dev`, the FQDN format of the hosts will be:
    - `example1.local.kyma.dev`
    - `example2.local.kyma.dev`


2. In order to identify the label key and value for the target workload check the selector from the service being exposed by those APIRules. 

    You can use this command:
    ```bash 
    kubectl get service <SERVICE_NAME> -n <NAMESPACE> -o jsonpath='{.spec.selector}'
    ```

   In this case, the selector for the `httpbin` service in the `default` namespace is `app: httpbin`:
    ```bash
   kubectl get service httpbin -n default -o jsonpath='{.spec.selector}'
    {"app":"httpbin"}%
    ```
   
3. Create a temporary AuthorizationPolicy using the gathered information in above points. This policy will allow traffic from the Ory Oathkeeper service and the Istio ingress gateway to the target workload only for the specified hosts.

    | Option                             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
    |------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
    | **{NAMESPACE}**                    | The namespace to which the AuthorizationPolicy applies. This namespace must include the target workload for which you allow traffic. The selector matches workloads in the same namespace as the AuthorizationPolicy.                                                                                                                                                                                                                                                                                                                |
    | **{HOSTNAME}**                     | List all host names that are exposed by the `v1beta1` APIRules being migrated which contain handler `no_auth` or `allow`. These hostnames represent the external endpoints through which clients access the workload. List all relevant hostnames in FQDN format to ensure the AuthorizationPolicy allows traffic for each endpoint during migration.                                                                                                                                                                                |
    | **{LABEL_KEY}**: **{LABEL_VALUE}** | To further restrict the scope of the AuthorizationPolicy, specify label selectors that match the target workload. Replace these placeholders with the actual key and value of the label. The label indicates a specific set of Pods to which a policy should be applied. The scope of the label search is restricted to the configuration namespace in which the AuthorizationPolicy is present. <br>For more information, see [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/). |

   This policy allows traffic from the ingress gateway to the target workload for all specified hosts and all traffic from the Ory Oathkeeper service. Applying this policy before starting the migration process ensures uninterrupted access to service endpoints during the migration.

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
    
     If you don't have any APIRules `v1beta1` with `allow` or `no_auth` handlers you can remove the second rule from the AuthorizationPolicy. The AuthorizationPolicy will then allow all traffic from the Ory Oathkeeper service to the target workload and look like following:

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
   
   Same case if you have APIRules `v1beta1` with only `allow` or `no_auth` handler you can specify only the second rule in the AuthorizationPolicy. The policy will then allow all traffic from the ingress gateway to the target workload for specified hosts.

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
                hosts:
                  - {HOSTNAME1}
                  - {HOSTNAME2}
      selector:
        matchLabels:
          {LABEL_KEY}: {LABEL_VALUE}
    ```

    In our example, the temporary applied AuthorizationPolicy would look like this as APIRules `v1beta1` have `no_auth` handler and `jwt` handler:

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
                  - example1.local.kyma.dev
                  - example2.local.kyma.dev
      selector:
        matchLabels:
          app: httpbin
    ```

4. Migrate APIRules `v1beta1` to `v2` following the steps in [migration guidelines](../apirule-migration/README.md). During this process, the temporary AuthorizationPolicy ensures that requests from both `v1beta1` APIRules to the target workload are allowed.

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
