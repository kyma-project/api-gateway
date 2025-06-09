# Retrieve the **spec** of APIRule in version `v1beta1`

This tutorial explains how to retrieve the originally applied **spec** of an APIRule in version `v1beta1` when the displayed **spec** appears empty in the Kyma dashboard and when using the `kubectl get` command.

## Context
APIRule version `v1beta1` is deprecated and scheduled for removal. Once the APIRule custom resource definition (CRD) stops serving version `v1beta1`, the API server will no longer respond to requests for APIRules in this version. Consequently, you will not be able to create, update, delete, or view APIRules in `v1beta1`. 

This creates a migration challenge: if your APIRule was originally created using `v1beta1` and you have not yet migrated to `v2`, you may find that the **spec** is empty when viewed in the Kyma dashboard or via the `kubectl get` command. 

In this situation, you must access the original APIRule `v1beta1` configuration through annotation. To learn how to do this, follow the procedure.

## Prerequisites

* You have a deployed workload with an APIRule in the deprecated `v1beta1` version.
* To use CLI instructions, you must have [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [yq](https://mikefarah.gitbook.io/yq). Alternatively, you can use Kyma dashboard.

## Procedure

<!-- tabs:start -->

#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules v1beta1** and select specific APIRule CR.
2. Go to **Edit** tab.
3. Copy the value of **metadata.annotations.gateway.kyma-project.io/v1beta1-spec** as it stores the original configuration of the APIRule created in `v1beta1`.


#### **kubectl**

To get the original **spec** of the APIRule created in version `v1beta1`, use the annotation that stores the original configuration. 

```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -ojsonpath='{.metadata.annotations.gateway\.kyma-project\.io/v1beta1-spec}' 
```
Sample output in JSON format:
```json
{"host":"httpbin.local.kyma.dev","service":{"name":"httpbin","namespace":"test","port":8000},"gateway":"kyma-system/kyma-gateway","rules":[{"path":"/anything","methods":["POST"],"accessStrategies":[{"handler":"noop"}]},{"path":"/.*","methods":["GET"],"accessStrategies":[{"handler":"allow"}]}]}
```
Format the output as YAML for better readability using the `yq` command.
```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -ojsonpath='{.metadata.annotations.gateway\.kyma-project\.io/v1beta1-spec}' | yq -P
```
Sample output in YAML format:
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
  - path: /.*
    methods:
      - GET
    accessStrategies:
      - handler: allow
```
<!-- tabs:end -->

Next adjust obtained configuration of the APIRule to migrate it to version `v2` by following the [Tutorials - Migrate APIRule from version `v1beta1` to version `v2`](./README.md). 