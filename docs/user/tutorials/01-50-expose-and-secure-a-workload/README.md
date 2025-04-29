# Tutorials - Expose and Secure a Workload
Browse the API Gateway tutorials to learn how to expose and secure workloads.

> [!WARNING]
> APIRule CR in version `v1beta1` has been deprecated and will be removed in upcoming releases. Version `v2alpha1`, introduced for testing purposes, will become deprecated after the stable APIRule `v2` is promoted to the regular channel.
> - After careful consideration, we have decided to postpone the release of API Gateway 3.0.0 (which contains the APIRules `v2` upgrade) to **May 5th**.
> - Additionally, the planned deletion of `v1beta1` for end of May will also be postponed. A new target date will be announced in the future.
> - **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`.
> 
> To migrate your APIRule CRs from version `v2alpha1` to version `v2`, you must update the version in APIRule CRs’ metadata.
> 
> To migrate your APIRule CRs from version `v1beta1` to version `v2`, follow the procedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). See [Changes Introduced in APIRule v2alpha1 and v2](https://help.sap.com/docs/link-disclaimer?site=https%3A%2F%2Fcommunity.sap.com%2Ft5%2Ftechnology-blogs-by-sap%2Fchanges-introduced-in-apirule-v2alpha1-and-v2%2Fba-p%2F14029529). 
> 
> Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure from version `v1beta1` to version `v2` is the same as from version `v1beta1` to version `v2alpha1`.

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