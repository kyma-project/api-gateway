# APIRule Custom Resource <!-- {docsify-ignore-all} -->

The `apirules.gateway.kyma-project.io` CRD describes the kind and the format of data the APIRule Controller uses to configure resources.

> [!WARNING]
> APIRule CRs in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases.
>
> After careful consideration, we have decided that the deletion of `v1beta1` planned for end of May will be postponed. A new target date will be announced in the future.
> 
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`.
> 
> To migrate your APIRule CRs from version `v2alpha1` to version `v2`, you must update the version in APIRule CRs’ metadata.
> 
> To migrate your APIRule CRs from version `v1beta1` to version `v2`, follow the procedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). See [Changes Introduced in APIRule v2alpha1 and v2](https://help.sap.com/docs/link-disclaimer?site=https%3A%2F%2Fcommunity.sap.com%2Ft5%2Ftechnology-blogs-by-sap%2Fchanges-introduced-in-apirule-v2alpha1-and-v2%2Fba-p%2F14029529). 
> 
> Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure from version `v1beta1` to version `v2` is the same as from version `v1beta1` to version `v2alpha1`.

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
