# Migrate APIRule `v1beta1` of type noop, allow or no_auth to version `v2`


This tutorial explains how to migrate an APIRule created with version `v1beta1` using the **noop**, **allow** or **no_auth** handler to version `v2`, where the **noAuth** handler replaces all of the above handlers from the `v1beta1` version.


## Context 

APIRule version `v1beta1` is deprecated and scheduled for removal. Once the APIRule custom resource definition (CRD) stops serving version `v1beta1`, the API server will no longer respond to requests for APIRules in this version. Consequently, you will not be able to create, update, delete, or view APIRules in `v1beta1`. Therefore, migrating to version `v2` is required.




## Prerequisites

* You have a deployed workload with the Istio and API Gateway modules enabled.
* To use the CLI instructions, you must have [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/) installed.
* You have obtained the configuration of the APIRule in version `v1beta1` to be migrated. See [Retrieve the **spec** of APIRule in version `v1beta1`](./01-81-retrieve-v1beta1-spec.md).
* The workload exposed by the APIRule in version `v2` must be a part of the Istio service mesh.

## Steps

> [!NOTE] In this example, the APIRule `v1beta1` was created with **noop**, **allow** and **no_auth** handlers, so the migration targets an APIRule `v2` using the **noAuth** handler. To illustrate the migration, the HTTPBin service is used, exposing the `/anything`, `/headers` and `/.*` endpoints. The HTTPBin service is deployed in its own namespace, with Istio enabled, ensuring the workload is part of the Istio service mesh.

1. Obtain a configuration of the APIRule in version `v1beta1` and save it for further modifications. For instructions, see [Retrieve the **spec** of APIRule in version `v1beta1`](./01-81-retrieve-v1beta1-spec.md). Below is a sample of the retrieved **spec** in YAML format for an APIRule in `v1beta1`:
```yaml
host: httpbin.local.kyma.dev
service:
  name: httpbin
  namespace: test
  port: 8000
gateway: kyma-system/kyma-gateway
rules:
  - path: /anything
    methods:
      - POST
    accessStrategies:
      - handler: noop
  - path: /headers
    methods:
      - HEAD
    accessStrategies:
      - handler: allow
  - path: /.*
    methods:
      - GET
    accessStrategies:
      - handler: no_auth
```
Above configuration uses the **noop** handler to expose `/anything`, the **allow** handler to expose `/headers` and the **no_auth** handler to expose `/.*` HTTPBin endpoints.

2. Adjust the obtained configuration to APIRule `v2` by replacing the **noop**, **allow** and **no_auth** handlers with the **noAuth** handler. This requires modifying the existing APIRule spec to ensure it is valid for the `v2` version with the **noAuth** type. Below is a sample of the adjusted APIRule in `v2`:
```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: httpbin
  namespace: test
spec:
  hosts:
    - httpbin
  service:
    name: httpbin
    namespace: test
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /anything
      methods: ["POST"]
      noAuth: true
    - path: /headers
      methods: ["HEAD"]
      noAuth: true      
    - path: /{**}
      methods: ["GET"]
      noAuth: true
```
> [!NOTE] 
> Notice that the **hosts** field accepts a short host name (without a domain). Additionally, the path `/.*` has been changed to `/{**}` because APIRule `v2` does not support regular expressions in the **spec.rules.path** field.  For more information about the changes introduced in APIRule `v2`, see the [APIRule v2 Changes](../../custom-resources/apirule/04-70-changes-in-apirule-v2.md) document. **Read this document before applying the new APIRule `v2`.**

3. Update the APIRule to version `v2` by applying the adjusted configuration. To verify the version of the applied APIRule, check the value of the `gateway.kyma-project.io/original-version` annotation in the APIRule spec. A value of `v2` indicates that the APIRule has been successfully migrated. You can use the following command:
```bash 
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
```
```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  annotations:
    gateway.kyma-project.io/original-version: v2
...
```
Above APIRule has been successfully migrated to version `v2`.

> [!WARNING] Do not manually change the `gateway.kyma-project.io/original-version` annotation. This annotation is automatically updated when you apply your APIRule in version `v2`.

4.To preserve the internal traffic policy from the APIRule `v1beta1`, you must apply the following AuthorizationPolicy. In APIRule `v2`, internal traffic is blocked by default. Without this AuthorizationPolicy, attempts to connect internally to the workload will result in an `RBAC: access denied` error. Ensure that the selector label is updated to match the target workload:

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: allow-internal
  namespace: ${NAMESPACE}
spec:
  selector:
    matchLabels:
      ${LABEL_KEY}: ${LABEL_VALUE} 
  action: ALLOW
  rules:
  - from:
    - source:
        notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
```

5. Additionally, to retain the CORS configuration from the APIRule `v1beta1`, update the APIRule in version `v2` to include the same CORS settings. For preflight requests work correctly, you must explicitly add the `"OPTIONS"` method to the **rules.methods** field of your APIRule `v2`. For guidance, refer to the available [APIRule `v2` samples](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/04-10-apirule-custom-resource?id=sample-custom-resource).

### Access Your Workload

- Send a `POST` request to the exposed workload:

  ```bash
  curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `HEAD` request to the exposed workload:

  ```bash
  curl -ik -I https://{SUBDOMAIN}.{DOMAIN_NAME}/headers
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `GET` request to the exposed workload:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/ip
  ```
  If successful, the call returns the `200 OK` response code.
