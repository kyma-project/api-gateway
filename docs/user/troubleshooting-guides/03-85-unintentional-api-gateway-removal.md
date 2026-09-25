# API Gateway Module Deletion Blocked

Follow the steps in this guide if the API Gateway module deletion is blocked because API Gateway resources still exist in the cluster.

## Symptom

The API Gateway custom resource (CR) is in the `Warning` state. The condition of type **Ready** is set to `false` with the reason `DeletionBlockedExistingResources`. To verify this, run:

```bash
kubectl get apigateway default -n kyma-system -o jsonpath='{.status.conditions[0]}'
```

You get an output similar to the following:

```bash
{"lastTransitionTime":"2026-03-20T10:25:31Z","message":"API Gateway deletion blocked because of the existing custom resources: apirule/multi-workload","reason":"DeletionBlockedExistingResources","status":"False","type":"Ready"}
```

## Cause

The API Gateway module wasn't completely removed because related resources still exist in the cluster.

For example, the issue occurs when you delete the API Gateway module, but there are still APIRule resources in the cluster. In such cases, the hooked finalizer pauses the deletion until you remove all the related resources. This [blocking deletion strategy](https://github.com/kyma-project/community/issues/765) is intentionally designed and is enabled by default for the API Gateway module.

## Solution

Choose one of the following options depending on whether you want to revert the deletion or permanently remove the API Gateway module.

### Revert an Accidental Deletion

If you unintentionally deleted the API Gateway module and want to restore the cluster to its previous state, remove the finalizers and re-add the module. The module reconciles back to a healthy state and all existing resources are preserved.

1. To edit the APIGateway CR, run:

    ```bash
    kubectl edit apigateway -n kyma-system default
    ```

2. Delete the following lines:

    ```yaml
    finalizers:
      - gateways.operator.kyma-project.io/api-gateway
      - gateways.operator.kyma-project.io/kyma-gateway
    ```

3. Save the changes.

4. Add the API Gateway module again. The APIGateway CR returns to the `Ready` state within a few seconds.

### Remove the API Gateway Module

If you intentionally deleted the API Gateway module, you must clean up the blocking resources yourself before the deletion completes.

1. To identify which resources are blocking the deletion, run:

    ```bash
    kubectl logs -n kyma-system deployments/api-gateway-controller-manager | grep "blocking deletion"
    ```

2. Remove the listed resources.

Once all blocking resources are removed, the API Gateway module deletion resumes automatically.
