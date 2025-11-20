# Tutorials

Learn how to set up a custom domain and Gateway:
- [Set Up a Custom Domain for a Workload](./01-10-setup-custom-domain-for-workload.md)
- [Set Up a TLS Gateway](./01-20-set-up-tls-gateway.md)
- [Configure mTLS Authentication in SAP BTP, Kyma Runtime](./01-10-mtls-authentication/configure-mtls-Gardener-certs.md)
- [Configure mTLS Authentication on k3d](./01-10-mtls-authentication/configure-mtls-k3d.md)

Learn how to expose and secure workloads:

> [!NOTE] 
> To expose a workload using APIRule in version `v2`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).

- [Expose and Secure a Workload with JWT Using SAP Cloud Identity Services](./01-40-expose-workload-jwt.md)
- [Expose a Workload with noAuth](./01-40-expose-workload-noauth.md) 
- [Secure a Workload with extAuth](./01-53-expose-and-secure-workload-ext-auth.md)
- [Expose Multiple Workloads](./01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces With a Single APIRule Definition](./01-42-expose-workloads-multiple-namespaces.md)
- [Use a Short Host Name](./01-43-expose-workload-short-host-name.md)
- [Configure IP-Based Access with XFF](./01-55-ip-based-access-with-xff.md)

Learn how to set up a custom identity provider:

- [Set Up a Custom Identity Provider](./01-60-security/01-62-set-up-idp.md)

Learn how to configure rate limit for a workload:

- [Set up local rate limits](./01-70-local-rate-limit.md)
