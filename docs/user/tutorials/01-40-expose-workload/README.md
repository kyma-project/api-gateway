# Tutorials - Expose a Workload
Browse the API Gateway tutorials to learn how to expose workloads. The tutorials are available in two versions: one uses the APIRule resource in version `v2alpha1` and the other uses the APIRule resource in version `v1beta1`. 

> [!WARNING]
> APIRule in version `v1beta1` has been deprecated. Version `v2alpha2` was introduced for testing purposes. It will become deprecated on March 31, 2025 and removed on June 16, 2025. The stable APIRule `v2` will be introduced on March 31, 2025. 
> 
> To migrate your APIRule CRs to version `v2`, follow the prcedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). Since the APIRule CRD `v2alpha2` is identical to `v2alpha2`, the migration procedure for both versions is the same. 
> 
> For more information on the timelines, see [APIRule migration - timelines](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-timelines/ba-p/13995712).

Expose a workload with APIRule in version `v2`:
- [Expose a Workload](./01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./01-42-expose-workloads-multiple-namespaces.md)

Expose a workload with APIRule in version `v2alpha1`:
- [Expose a Workload](./v2alpha1/01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./v2alpha1/01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./v2alpha1/01-42-expose-workloads-multiple-namespaces.md)

Expose a workload with APIRule in version `v1beta1`:
- [Expose a Workload](./v1beta1/01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./v1beta1/01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./v1beta1/01-42-expose-workloads-multiple-namespaces.md)
