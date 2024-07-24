# APIRule Contains an Empty Spec

## Symptoms

- There is an empty spec in an `APIRule`, for example:
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

The conversion from the `APIRule` version `v1beta1` to `v2alpha1` is not possible. It's only possible to convert `jwt` handler from `v1beta1` to `v2alpha1`, for more information see the [documentation](https://github.com/kyma-project/api-gateway/blob/main/docs/user/custom-resources/apirule/v2alpha1/04-60-apirule-migration.md).

## Remedy

Get the APIRule in its original version:
  ```bash
  kubectl get apirules.v1beta1.gateway.kyma-project.io -n NAMESPACE APIRULENAME -oyaml
  ```
