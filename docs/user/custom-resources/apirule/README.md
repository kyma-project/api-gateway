# APIRule Custom Resource <!-- {docsify-ignore-all} -->

The `apirules.gateway.kyma-project.io` CRD describes the kind and the format of data the APIRule Controller uses to configure resources.

> [!WARNING]
> APIRule CR in version `v1beta1` has been deprecated and will be removed in upcoming releases. Version `v2alpha1`, introduced for testing purposes, will become deprecated after the stable APIRule `v2` is promoted to the regular channel. The promotion of the APIRule `v2` to the regular channel has been postponed. We will keep you posted on the coming dates and changes.
> 
> To migrate your APIRule CRs to version `v2`, follow the procedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure for both versions is the same.

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
