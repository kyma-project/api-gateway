# Issues with APIRules and Service Connection

If you have issues creating an APIRule custom resource (CR) or you've exposed a service but you cannot connect to it, see the troubleshooting guides related to:

- [APIRule CR in version v2](./03-00-basic-diagnostics.md)
- [APIRule CR in version v2alpha1](./v2alpha1/03-00-basic-diagnostics.md)
- [APIRule CR in version v1beta1](./03-00-basic-diagnostics.md)

> [!WARNING]
> APIRule in version `v1beta1` has been deprecated. Version `v2alpha1` was introduced for testing purposes. It will become deprecated on March 31, 2025 and removed on June 16, 2025. The stable APIRule `v2` will be introduced on March 31, 2025. 
> 
> To migrate your APIRule CRs to version `v2`, follow the prcedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure for both versions is the same. 
> 
> For more information on the timelines, see [APIRule migration - timelines](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-timelines/ba-p/13995712).