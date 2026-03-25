# Reverting the Istio Module's Deletion
Follow the steps outlined in this troubleshooting guide if you unintentionally deleted the Istio module and want to restore the cluster to its normal state without losing any resources created in the cluster.

## Symptom

The API Gateway custom resource (CR) is in the `Warning` state. The condition of type **Ready** is set to `false` with the reason `DeletionBlockedExistingResources`. To verify this, run the command:

```bash
kubectl get istio default -n kyma-system -o jsonpath='{.status.conditions[0]}'
```

You get an output similar to this one:

```bash
{"lastTransitionTime":"2026-03-20T10:25:31Z","message":"API Gateway deletion blocked because of the existing custom resources: apirule/multi-workload","reason":"DeletionBlockedExistingResources","status":"False","type":"Ready"}
```

>### Note:
> If you intended to delete the API Gateway module, the symptoms described in this document are expected, and you must clean up the remaining resources yourself. To check which resources are blocking the deletion, see the logs of the `api-gateway-controller-manager` container.

## Cause

The API Gateway module wasn't completely removed because related resources still exist in the cluster.

For example, the issue occurs when you delete the API Gateway module, but there are still APIRule resources in the cluster. In such cases, the hooked finalizer pauses the deletion of the API Gateway module until you remove all the related resources. This [blocking deletion strategy](https://github.com/kyma-project/community/issues/765) is intentionally designed and is enabled by default for the module.


## Solution

1. To edit the APIGateway CR, run:
    ```bash
    kubectl edit istio -n kyma-system default
    ```
2. To remove the finalizers from the APIGateway CR, delete the following lines:
    ```bash
    finalizers:
      - gateways.operator.kyma-project.io/api-gateway
      - gateways.operator.kyma-project.io/kyma-gateway
    ```
    When the finalizers are removed, the API Gateway module is deleted. All the other resources remain in the cluster.
3. Save the changes.
4. Add the API Gateway module again.

When you re-add the API Gateway module, its reconciliation is reinitiated. The API Gateway CR returns to the `Ready` state within a few seconds.