# APIRule `v2` Contains an Empty Spec

## Symptoms
There is an empty **spec** in an APIRule custom resource (CR), for example:

  ```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
  ```yaml
  apiVersion: gateway.kyma-project.io/v2
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

The conversion from the APIRule CR in version `v1beta1` to version `v2` is not possible. 
It's only possible to convert the `noAuth` and `jwt` handlers from `v1beta1` to `v2`. 
Beware that the `jwt` handler with more than one `trusted_issuers` or `jwks_urls` also cannot be converted.

## Solution

Get the APIRule in its original version:
  ```bash
  kubectl get apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```

