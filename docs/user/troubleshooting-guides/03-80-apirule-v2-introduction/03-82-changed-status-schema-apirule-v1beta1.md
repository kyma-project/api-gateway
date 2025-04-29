# APIRule `v1beta1` Contains a Changed Status Schema

## Symptoms
There is a changed schema of **status** in an APIRule custom resource (CR), for example:


  ```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
  ```yaml
  status:
    lastProcessedTime: "2025-04-25T11:16:11Z"
    state: Ready
  ```

## Cause
The conversion from the APIRule CR in version `v1beta1` to version `v2` is not possible. 
Schema of the `status.state` field in the `v2` APIRule custom resource introduces unified approach as in the API Gateway custom resource.
The possible states of **status.state** field are  `Ready`, `Warning`, `Error`, `Processing`, or `Deleting`.

## Solution

Get the APIRule in its original version:
  ```bash
  kubectl get apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
```yaml
  status:
  APIRuleStatus:
    code: OK
  accessRuleStatus:
    code: OK
  lastProcessedTime: "2025-04-25T11:16:11Z"
  observedGeneration: 1
  virtualServiceStatus:
    code: OK  
```