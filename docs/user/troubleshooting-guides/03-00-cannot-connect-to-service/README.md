# Issues with APIRules and Service Connection

If you have issues creating an APIRule custom resource (CR) or you've exposed a service but you cannot connect to it, see the troubleshooting guides related to:

- [APIRule CR in version `v2`](./03-00-basic-diagnostics.md)
- [APIRule CR in version `v2alpha1`](./v2alpha1/03-00-basic-diagnostics.md)
- [APIRule CR in version `v1beta1`](./03-00-basic-diagnostics.md)

> [!WARNING]
> APIRule CRDs in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases. Due to the upcoming deletion, managing APIRules `v1beta1` using Kyma dashboard is no longer possible. Additionally, you can't create APIRules `v1beta1` in new clusters. For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).
> 
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md).