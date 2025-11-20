# Unable to Create, Modify, or Delete an APIRule `v1beta1`

## Symptoms
- Kyma dashboard does not display APIRules created in version `v1beta1`.

- APIRules obtained via `kubectl get` are in the `Warning` status.
  ```bash
  kubectl get apirules.gateway.kyma-project.io -A
  
  NAMESPACE   NAME                STATUS    HOSTS
  default     example-apirule     Warning   ["example-host"]
  ```
  
- When you try to create, modify, or delete an APIRule created in version `v1beta1` using kubectl, you encounter an error related to the admission webhook.
  ```bash
  kubectl apply apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  
  Warning: Version v1beta1 of APIRule is deprecated and will be removed in future releases. Use version v2 instead.
  Error from server (Forbidden): error when creating "STDIN": admission webhook "v1beta1-admission.apirule.gateway.kyma-project.io" denied the request: v1beta1 APIRule version is no longer supported, please use v2 instead
  ```
  ```bash
  kubectl edit apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  
  Warning: Version v1beta1 of APIRule is deprecated and will be removed in future releases. Use version v2 instead.
  error: apirules.gateway.kyma-project.io "hello-kymav1beta1" could not be patched: admission webhook "v1beta1-admission.apirule.gateway.kyma-project.io" denied the request: v1beta1 APIRule version is no longer supported, please use v2 instead
  You can run `kubectl replace -f <temporary-file-path>` to try this update again.
  ```

  ```bash
  kubectl delete apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml

  Warning: Version v1beta1 of APIRule is deprecated and will be removed in future releases. Use version v2 instead.
  Error from server (Forbidden): admission webhook "v1beta1-admission.apirule.gateway.kyma-project.io" denied the request: v1beta1 APIRule version is no longer supported, please use v2 instead
  ```

## Cause
The APIRule custom resource `v1beta1` is deleted in version 3.4 of the API Gateway module. As a result, changes have been introduced to begin migration to the latest stable version, `v2`. While all `v1beta1` APIRules remain fully operational in the background, you can't create, modify, or delete APIRule CRs in version `v1beta1`. Additionally, you can't display APIRules `v1beta1` in Kyma dashboard. Only `v2` APIRules are supported, as `v2` is now the latest stable APIRule version in the Kubernetes API.

However, all `v1beta1` APIRule configurations created before this change in existing clusters remain active and continue to function as expected. The API Gateway module manages and reconciles resources based on the existing `v1beta1` APIRule settings.


> [!NOTE]
>  For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).

## Solution

To make sure that support for your APIRules is maintained, you must migrate them to version `v2`.
To learn how to do this, follow the [APIRule migration guide](../apirule-migration/README.md).

