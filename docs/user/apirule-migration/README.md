# Migrate APIRule from Version `v1beta1` to Version `v2`
APIRule custom resource (CR) `v1beta1` has been deprecated and scheduled for deletion. You must migrate all your APIRule CRs to version `v2`. Learn more about the timeline and see how to perform the migration.


To migrate to version v2, follow the steps:

1. To identify which APIRules must be migrated, run the following command:
    ```bash
    kubectl get apirules.gateway.kyma-project.io -A -o json | jq '.items[] | select(.metadata.annotations["gateway.kyma-project.io/original-version"] == "v1beta1") | {namespace: .metadata.namespace, name: .metadata.name}'
    ```

2. To retrieve the complete **spec** with the rules field of an APIRule in version `v1beta1`, see [Retrieving the Complete **spec** of an APIRule in Version `v1beta1`](./01-81-retrieve-v1beta1-spec.md).

3. To migrate an APIRule from version `v1beta1` to version `v2`, follow the relevant guide:
    - [Migrating APIRule v1beta1 of Type jwt to Version v2](./01-83-migrate-jwt-v1beta1-to-v2.md)
    - [Migrating APIRule v1beta1 of Type noop, allow, or no_auth to Version v2](./01-82-migrate-allow-noop-no_auth-v1beta1-to-v2.md)
    - [Migrating APIRule v1beta1 of type oauth2_introspection to version v2](./01-84-migrate-oauth2-v1beta1-to-v2.md)

For more information about APIRule v2, see also [APIRule `v2` Custom Resource](../custom-resources/apirule/04-10-apirule-custom-resource.md) and [Changes Introduced in APIRule `v2`](../custom-resources/apirule/04-70-changes-in-apirule-v2.md).

For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/whats-new/cf0cb2cb149647329b5d02aa96303f56?locale=en-US&Component=Kyma+Runtime&Valid_as_Of=2025-08-12:2025-08-12).