# APIRule Custom Resource <!-- {docsify-ignore-all} -->

The `apirules.gateway.kyma-project.io` CRD describes the kind and the format of data the APIRule Controller uses to configure resources.

> [!WARNING]
> APIRule CRDs in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases. Due to the upcoming deletion, managing APIRules `v1beta1` using Kyma dashboard is no longer possible. Additionally, you can't create APIRules `v1beta1` in new clusters. For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).
> 
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md).

Browse the documentation related to the APIRule CR in version `v2`:
- [Specification of APIRule CR](./04-10-apirule-custom-resource.md)
- [APIRule Access Strategies](./04-15-api-rule-access-strategies.md)

Browse the documentation related to the APIRule CR in version `v2alpha1`:
- [Specification of APIRule CR](./v2alpha1/04-10-apirule-custom-resource.md)
- [APIRule Access Strategies](./v2alpha1/04-15-api-rule-access-strategies.md)

Browse the documentation related to the APIRule CR in version `v1beta1`:
- [Specification of APIRule CR](./v1beta1-deprecated/04-10-apirule-custom-resource.md)
- [Istio JWT Access Strategy](./v1beta1-deprecated/04-20-apirule-istio-jwt-access-strategy.md)
- [Comparison of Ory Oathkeeper and Istio JWT Access Strategies](./v1beta1-deprecated/04-30-apirule-jwt-ory-and-istio-comparison.md)
- [APIRule Mutators](./v1beta1-deprecated/04-40-apirule-mutators.md)
- [OAuth2 and JWT Authorization](./v1beta1-deprecated/04-50-apirule-authorizations.md)
