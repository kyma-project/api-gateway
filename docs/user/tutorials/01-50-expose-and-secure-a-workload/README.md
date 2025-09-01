# Tutorials - Expose and Secure a Workload
Browse the API Gateway tutorials to learn how to expose and secure workloads.

> [!WARNING]
> APIRule CRDs in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases. Due to the upcoming deletion, managing APIRules `v1beta1` using Kyma dashboard is no longer possible. Additionally, you can't create APIRules `v1beta1` in new clusters. For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US&version=Cloud#apirule-v1beta1-migration-timeline).
> 
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md).

> [!NOTE] 
> To expose a workload using APIRule in version `v2` or `v2alpha1`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).

Browse the API Gateway tutorials to learn how to expose workloads.

- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with extAuth](./01-53-expose-and-secure-workload-ext-auth.md)
- [Expose and Secure a Workload with a Certificate](./01-54-expose-and-secure-workload-with-certificate.md)
- [Use the XFF Header to Configure IP-Based Access to a Workload](./01-55-ip-based-access-with-xff.md)