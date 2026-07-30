# APIRule `v1beta1` Is No Longer Reconciled

## Symptoms

- An APIRule created in version `v1beta1` is in the `Error` status.
  ```bash
  kubectl get apirules.gateway.kyma-project.io -A

  NAMESPACE   NAME                STATUS   HOSTS
  default     example-apirule     Error    ["example-host"]
  ```

- The status description of the affected APIRule contains the following message:
  ```
  Version v1beta1 of APIRule is no longer supported and this APIRule is not reconciled. Make sure to migrate to version v2.
  ```

- Changes to the APIRule `v1beta1` configuration, or to its sub-resources (VirtualService, AuthorizationPolicy, RequestAuthentication), are no longer applied by the API Gateway module.

## Cause

Reconciliation of APIRule `v1beta1` is disabled. The API Gateway module no longer reconciles APIRules created in version `v1beta1`, and it no longer owns or manages their sub-resources. As a result:

- Any modification to a `v1beta1` APIRule is not reconciled back.
- Any modification or deletion of the sub-resources created by the APIRule controller is not reverted.
- Deletion of a `v1beta1` APIRule is not handled by the controller.

Workloads that were already exposed remain exposed and secured, because the existing sub-resources are left in place. However, because they are no longer reconciled, any change to them might cause disruption in the availability of, or access to, the exposed workloads.

> [!NOTE]
> For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).

## Solution

To restore reconciliation and make sure that support for your APIRules is maintained, migrate them to version `v2`.
To learn how to do this, follow the [APIRule migration guide](../apirule-migration/README.md).
