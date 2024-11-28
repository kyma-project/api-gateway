# APIRule Contains an Empty **spec**

## Symptoms

There is an empty **spec** in an APIRule custom resource (CR), for example:
  ```yaml
  apiVersion: gateway.kyma-project.io/v2alpha1
  kind: APIRule
  metadata:
    name: httpbin
    namespace: test
  spec: {}
  status:
    lastProcessedTime: "2024-07-22T07:06:59Z"
    state: Ready
  ```

## Cause

The conversion from the APIRule CR in version `v1beta1` to version `v2alpha1` is not possible. It's only possible to convert the `noAuth` and `jwt` handlers from `v1beta1` to `v2alpha1`. Beware that the `jwt` handler with more than one `trusted_issuers` or `jwks_urls` also cannot be converted.

## Solution

Get the APIRule in its original version:
  ```bash
  kubectl get apirules.v1beta1.gateway.kyma-project.io -n NAMESPACE APIRULENAME -oyaml
  ```
