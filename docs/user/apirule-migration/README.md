# Migrate APIRule from Version `v1beta1` to Version `v2`
APIRule custom resource (CR) `v1beta1` is deleted. You must migrate all your APIRule CRs to version `v2`. Learn how to perform the migration.

> [!WARNING]
> APIRule CRD `v2` is the latest stable version. Version `v1beta1` is removed in release 3.4 of the API Gateway module. 
>- You can no longer create, edit, or delete APIRules `v1beta1`. All existing `v1beta1` APIRule configurations continue to function as expected. To make any changes, migrate to version `v2`.
>- Reconciliation and migration of APIRules `v1beta1` will be disabled with release 3.9. Once reconciliation is disabled, APIRules `v1beta1` will continue to function as currently configured, but the API Gateway module will no longer own or manage them.
>- APIRules `v1beta1` are no longer visible in the Kyma dashboard. You can still display them using kubectl, but the resources are displayed in the converted `v2` format.

>
> **Required action**: Migrate all your APIRule CRs to version `v2`. For the complete timeline for SAP BTP, Kyma runtime, follow [API Gateway What's New notes](https://help.sap.com/whats-new/cf0cb2cb149647329b5d02aa96303f56?locale=en-US&version=Cloud&q=API+Gateway+module:).

To migrate to version v2, follow the steps:

1. To identify which APIRules must be migrated, run the following command:
    ```bash
    kubectl get apirules.gateway.kyma-project.io -A -o json | jq '.items[] | select(.metadata.annotations["gateway.kyma-project.io/original-version"] == "v1beta1") | {namespace: .metadata.namespace, name: .metadata.name}'
    ```

2. If two or more of your APIRules target the same workload, apply an additional AuthorizationPolicy to avoid traffic disruption during migration. See [Migrate Multiple APIRules Targeting the Same Workload from `v1beta1` to `v2`](./01-90-migrate-multiple-apirules-targeting-same-workload.md).

3. To retrieve the complete **spec** with the rules field of an APIRule in version `v1beta1`, see [Retrieving the Complete **spec** of an APIRule in Version `v1beta1`](./01-81-retrieve-v1beta1-spec.md).

4. To migrate an APIRule from version `v1beta1` to version `v2`, follow the relevant guide:
    - [Migrating APIRule v1beta1 of Type jwt to Version v2](./01-83-migrate-jwt-v1beta1-to-v2.md)
    - [Migrating APIRule v1beta1 of Type noop, allow, or no_auth to Version v2](./01-82-migrate-allow-noop-no_auth-v1beta1-to-v2.md)
    - [Migrating APIRule v1beta1 of type oauth2_introspection to version v2](./01-84-migrate-oauth2-v1beta1-to-v2.md)

For more information about APIRule `v2`, see also [APIRule `v2` Custom Resource](../custom-resources/apirule/04-10-apirule-custom-resource.md) and [Changes Introduced in APIRule `v2`](../custom-resources/apirule/04-70-changes-in-apirule-v2.md).