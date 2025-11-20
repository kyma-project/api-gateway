# Migrate APIRule from Version `v1beta1` to Version `v2`
APIRule custom resource (CR) `v1beta1` has been deprecated and scheduled for deletion. You must migrate all your APIRule CRs to version `v2`. Learn more about the timeline and see how to perform the migration.

> [!WARNING]
> APIRule CRD `v2` is the latest stable version. Version `v1beta1` is removed in release 3.4 of the API Gateway module. All existing `v1beta1` APIRule configurations continue to function as expected, but are not visible in Kyma dashboard. You can display APIRules  `v1beta1` using kubectl, but you can no longer edit them or create new APIRules `v1beta1`. To make changes, you must migrate the APIRule to version `v2`. For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US#apirule-v1beta1-migration-timeline).
>
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md).


## How to Migrate APIRules to Version v2

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

## APIRule v1beta1 Migration Timeline

APIRule CRDs in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases. Due to the upcoming deletion, managing APIRules `v1beta1` using Kyma dashboard is no longer possible. Additionally, you can't create APIRules `v1beta1` in new clusters.

For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).