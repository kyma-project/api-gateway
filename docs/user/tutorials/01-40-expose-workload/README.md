# Tutorials - Expose a Workload
Browse the API Gateway tutorials to learn how to expose workloads. The tutorials are available for the following versions of the APIRule custom resource (CR): `v2`, `v2alpha1`, and `v1beta1`. 

> [!WARNING]
> APIRule CRD `v2` is the latest stable version. Version `v1beta1` has been deprecated and will be removed in upcoming releases. Due to the upcoming deletion, managing APIRules `v1beta1` using Kyma dashboard is no longer possible. Additionally, you can't create APIRules `v1beta1` in new clusters. For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).
>
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md).

> [!NOTE] 
> To expose a workload using APIRule in version `v2` or `v2alpha1`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).

Browse the API Gateway tutorials to learn how to expose and secure workloads:
- [Expose a Workload](./01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./01-42-expose-workloads-multiple-namespaces.md)