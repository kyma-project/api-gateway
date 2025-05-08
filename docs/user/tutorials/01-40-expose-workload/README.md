# Tutorials - Expose a Workload
Browse the API Gateway tutorials to learn how to expose workloads. The tutorials are available in two versions: one uses the APIRule resource in version `v2alpha1` and the other uses the APIRule resource in version `v1beta1`. 

> [!WARNING]
> APIRule CR in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases.
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

Expose a workload with APIRule in version `v2`:
- [Expose a Workload](./01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./01-42-expose-workloads-multiple-namespaces.md)

Expose a workload with APIRule in version `v2alpha1`:
- [Expose a Workload](./v2alpha1/01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./v2alpha1/01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./v2alpha1/01-42-expose-workloads-multiple-namespaces.md)

Expose a workload with APIRule in version `v1beta1`:
- [Expose a Workload](./v1beta1-deprecated/01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./v1beta1-deprecated/01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./v1beta1-deprecated/01-42-expose-workloads-multiple-namespaces.md)
