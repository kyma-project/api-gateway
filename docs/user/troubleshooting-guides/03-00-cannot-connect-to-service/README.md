# Issues with APIRules and Service Connection

If you have issues creating an APIRule custom resource (CR) or you've exposed a service but you cannot connect to it, see the troubleshooting guides related to:

- [APIRule CR in version `v2`](./03-00-basic-diagnostics.md)
- [APIRule CR in version `v2alpha1`](./v2alpha1/03-00-basic-diagnostics.md)
- [APIRule CR in version `v1beta1`](./03-00-basic-diagnostics.md)

> [!WARNING]
> APIRule CR in version `v1beta1` has been deprecated and will be removed on May 12, 2025. Version `v2alpha1`, introduced for testing purposes, will become deprecated after the stable APIRule `v2` is promoted to the regular channel. The promotion of the APIRule `v2` to the regular channel has been postponed. We will keep you posted on the coming dates and changes.
> 
> Upon receiving customer feedback, we decided to postpone the promotion of the API Gateway module to the regular channel. To ensure the highest quality and reliability of our product, we want to take the necessary time to thoroughly investigate and resolve the matter. We will keep you posted on the coming dates and changes. We appreciate your understanding and patience.
> 
> To migrate your APIRule CRs to version `v2`, follow the prcedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure for both versions is the same.