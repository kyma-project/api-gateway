# Previously created APIRules are missing from Kyma Dashboard API Rules view

In case you have created APIRules using the `v1beta1` version, and you have not yet migrated them to the `v2` version,
they will not be displayed in the Kyma Dashboard API Rules view.
To display the APIRules in the Kyma Dashboard, you need to migrate them to the `v2` version.

## Symptom

You have created APIRules using the `v1beta1` version, and they are not displayed in the Kyma Dashboard API Rules view.
To check if the APIRules are present in the cluster, you can run the following command:

```bash
kubectl get apirules.v2.gateway.kyma-project.io -A
```
If the APIRules are present, you will see them listed in the output.
However, for APIRules that were not upgraded to the `v2` version, the output will not show the rules field, for example:

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  annotations:
    gateway.kyma-project.io/original-version: v1beta1
  name: httpbin
  namespace: test
spec:
    gateway: kyma-system/kyma-gateway
    hosts:
        - httpbin.local.kyma.dev
    service:
        name: httpbin
        namespace: test
        port: 8000
status:
    lastProcessedTime: "2025-04-25T11:16:11Z"
    state: Warning
```

In case the APIRule was created using the `v1beta1` version, the output will show 
the `gateway.kyma-project.io/original-version: v1beta1` annotation.

## Cause

The APIRules were originally created using the `v1beta1` version, and you haven't yet migrated them to the `v2` version.
The APIRule v1beta1 API is no longer available either via the Kyma Dashboard or the `kubectl` command.
To make sure that the support for the APIRule is maintained, you need to migrate the APIRules to the `v2` version.

## Solution

To migrate the APIRules to the `v2` version, you can follow the [APIRule migration guide](../../apirule-migration/README.md).
