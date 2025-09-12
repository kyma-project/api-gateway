# APIRule Custom Resource <!-- {docsify-ignore-all} -->

The `apirules.gateway.kyma-project.io` CRD describes the kind and the format of data the APIRule Controller uses to configure resources.

> [!WARNING]
> APIRule CRD `v2` is the latest stable version. Version `v1beta1` has been deprecated and will be removed in upcoming releases. All previously existing `v1beta1` APIRule configurations will continue to function as expected but will not be visible in the Kyma Dashboard due to the deletion. Management through the Kyma Dashboard or `kubectl` is not possible. To make changes, you must migrate the APIRule to latest stable version `v2`. Additionally, you can't create APIRules `v1beta1` in new clusters. For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).
> 
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md).

Browse the documentation related to the APIRule CR in version `v2`:
- [Specification of APIRule CR](./04-10-apirule-custom-resource.md)
- [APIRule Access Strategies](./04-15-api-rule-access-strategies.md)
