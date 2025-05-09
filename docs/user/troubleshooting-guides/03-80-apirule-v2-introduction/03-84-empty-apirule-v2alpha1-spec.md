# APIRule `v2alpha1` Contains an Empty Spec

## Symptom
There is an empty **spec** in an APIRule custom resource (CR), for example:

  ```bash
kubectl get apirules.v2alpha1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
  ```yaml
  apiVersion: gateway.kyma-project.io/v2aplha1
  kind: APIRule
  metadata:
    name: httpbin
    namespace: test
  spec: {}
  status:
    lastProcessedTime: "2025-04-25T11:16:11Z"
    state: Ready
  ```

## Cause

The APIRule was originally created using version `v1beta1`, and you haven't yet migrated it neither to version `v2alpha1` nor `v2`. When you try to display it specifying version `v2alpha1` in the command, a conversion from `v1beta1` to `v2alpha1` is performed. This conversion only affects the displayed resource's textual format and does not modify the resource in the cluster. 

If the full conversion is possible, the **spec** is presented in the output. 
However, if the conversion cannot be fully completed, the **spec** appears empty, and the original **spec** is stored in the resource's annotations.

## Solution

Specify explicitly `v1beta1` version when requesting the APIRule resource:
  ```bash
  kubectl get apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
