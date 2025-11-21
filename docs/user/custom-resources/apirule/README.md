# APIRule Custom Resource <!-- {docsify-ignore-all} -->

The `apirules.gateway.kyma-project.io` CRD describes the kind and the format of data the APIRule Controller uses to configure resources.

> [!WARNING]
> APIRule CRD `v2` is the latest stable version. Version `v1beta1` is removed in release 3.4 of the API Gateway module. 
>- All existing `v1beta1` APIRule configurations continue to function as expected.
>- APIRules v1beta1 are no longer visible in the Kyma dashboard. You can still display them using kubectl, but the resources are displayed in the converted v2 format.
>- You can no longer create new v1beta1 APIRules or edit the existing ones. To make any changes, you must migrate to version v2.
>
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md). For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US#apirule-v1beta1-migration-timeline).

Browse the documentation related to the APIRule CR in version `v2`:
- [Specification of APIRule CR](./04-10-apirule-custom-resource.md)
- [APIRule Access Strategies](./04-15-api-rule-access-strategies.md)
