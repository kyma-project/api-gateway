# Tutorials - Expose and Secure a Workload
Browse the API Gateway tutorials to learn how to expose and secure workloads.

> [!WARNING]
> APIRule CR in version `v1beta1` has been deprecated and will be removed in upcoming releases. Version `v2alpha1`, introduced for testing purposes, will become deprecated after the stable APIRule `v2` is promoted to the regular channel. The promotion of the APIRule `v2` to the regular channel has been postponed. We will keep you posted on the coming dates and changes.
> 
> To migrate your APIRule CRs to version `v2`, follow the procedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure for both versions is the same.

Expose and secure a workload with APIRule in version `v2`:
- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with extAuth](./01-53-expose-and-secure-workload-ext-auth.md)

Expose and secure a workload with APIRule in version `v2alpha1`:
- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./v2alpha1/01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with extAuth](./v2alpha1/01-53-expose-and-secure-workload-ext-auth.md)

Expose and secure a workload with APIRule in version `v1beta1`:
- [Expose and Secure a Workload with OAuth2](./v1beta1-deprecated/01-50-expose-and-secure-workload-oauth2.md)
- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./v1beta1-deprecated/01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with Istio](./v1beta1-deprecated/01-53-expose-and-secure-workload-istio.md)

[Expose and Secure a Workload with a Certificate](./01-54-expose-and-secure-workload-with-certificate.md)

[Use the XFF Header to Configure IP-Based Access to a Workload](./01-55-ip-based-access-with-xff.md)