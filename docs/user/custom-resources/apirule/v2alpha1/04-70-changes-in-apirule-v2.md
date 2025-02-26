# Changes Introduced in APIRule v2alpha1 and v2

This document presents all significant changes that APIRule `v2alpha1` introduces. Since version `v2alpha1` is identical to the stable version `v2`, you must consider these changes when migrating either to version `v2` or `v2alpha1`.

See the changes introduced in new versions:
- [A Workload Must Be in the Istio Service Mesh](#a-workload-must-be-in-the-istio-service-mesh)
- [Internal Traffic to Workloads Is Blocked by Default](#internal-traffic-to-workloads-is-blocked-by-default)
- [CORS Policy Is Not Applied by Default](#cors-policy-is-not-applied-by-default)
- [Path Specification Must Not Contain Regexp](#path-specification-must-not-contain-regexp)
- [JWT Configuration Requires Explicit Issuer URL](#jwt-configuration-requires-explicit-issuer-url)
- [Oathkeeper Removal](#oathkeeper-removal)
  - [Removed Support for Oathkeeper OAuth2 Handlers](#removed-support-for-oathkeeper-oauth2-handlers)
  - [Removed Support for Oathkeeper Mutators](#removed-support-for-oathkeeper-mutators)
- [Removed Support for Opaque Tokens](#removed-support-for-opaque-tokens)

> [!WARNING]
> APIRule CR in version `v1beta1` has been deprecated and will be removed on May 12, 2025. Version `v2alpha1`, introduced for testing purposes, will become deprecated on March 31, 2025 and removed on June 16, 2025. The stable APIRule `v2` is planned to be introduced on March 31, 2025, in the regular channel.
> 
> To migrate your APIRule CRs to version `v2`, follow the procedure described in the blog posts [APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833) and [APIRule migration - Ory Oathkeeper based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184). Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure for both versions is the same. 
> 
> For more information on the timelines, see [APIRule migration - timelines](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-timelines/ba-p/13995712).

## A Workload Must Be in the Istio Service Mesh

To use APIRules in versions `v2` or `v2alpha1`, the workload that an APIRule exposes must be in the Istio service mesh. If the workload is not inside the Istio service mesh, the APIRule will not work as expected.

**Required action**: To add a workload to the Istio service mesh, [enable Istio sidecar proxy injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection).

## Internal Traffic to Workloads Is Blocked by Default

By default, the access to the workload from internal traffic is blocked. This approach aligns with Kyma's principle of being "secure by default". In one of the future releases of the API Gateway module, the APIRule CR will contain a new field **internalTraffic** set to `Deny` by default. This field will allow you to permit traffic from the CR. For more information on this topic, see issue [#1632](https://github.com/kyma-project/api-gateway/issues/1632).

## CORS Policy Is Not Applied by Default

Version v1beta1 applied the following CORS onfiguration by default:
```yaml
Access-Control-Allow-Origins: "*"
Access-Control-Allow-Methods: "GET,POST,PUT,DELETE,PATCH"
Access-Control-Allow-Headers: "Authorization,Content-Type,*"
```

Versions `v2` and `v2alpha1` do not apply these default values. If the **corsPolicy** field is empty, the CORS configuration is not applied. For more information, see [architecture decision record #752](https://github.com/kyma-project/api-gateway/issues/752).

**Required action**: If you want to use default CORS values defined in `v1beta1` APIRule, you must explicitly define them in **corsPolicy** field.

## Path Specification Must Not Contain Regexp

APIRule v2alpha1 does not support regexp in the **spec.rules.path** field of APIRule CR. Instead, it supports use of the `{*}` and `{**}` operators. See the supported configurations:
- Use the exact path (for example, `/abc`). It matches the specified path exactly.
- Use the `{*}` operator (for example, `/foo/{*}` or `/foo/{*}/bar`).  This operator represents any request that matches the given pattern, with exactly one path segment replacing the operator.
- Use the `{**}` operator (for example, `/foo/{**}` or `/foo/{**}/bar`). This operator represents any request that matches the pattern with zero or more path segments in the operator’s place. It must be the last operator in the path.
- Use the wildcard path `/*`, which matches all paths. It’s equivalent to the `/{**}` path. If your configuration in APIRule `v1beta1` used such a path as `/foo(.*)`, when migrating to the new versions, you must define configurations for two spearate paths: `/foo` and `/foo/{**}`.

For more information on the APIRule specification, see [APIRule v2alpha1 Custom Resource](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-10-apirule-custom-resource).

**Required action**: Replace regexp expressions in the **spec.rules.path** field of your APIRule CRs with the `{*}` and `{**}` operators.

## JWT Configuration Requires Explicit Issuer URL

Versions `v2` and `v2alpha1` of APIRule introduce an additional mandatory configuration filed for JWT-based authorization - **issuer**. You must provide explicit issuer URL in the APIRule CR. See an example configuration:

```yaml
rules:
- jwt:
    authentications:
        -   issuer: {YOUR_ISSUER_URL}
            jwksUri: {YOUR_JWKS_URI}
```
If you use Cloud Identity Services, you can find the issuer URL in the OIDC well-known configuration at `https://{YOUR_TENANT}.accounts.ondemand.com/.well-known/openid-configuration`.

**Required action**: Add the **issuer** field to your APIRule specification when migrating to the new version. For more information on migration procedure for the `jwt` handler, see [SAP BTP, Kyma runtime: APIRule migration - noAuth and jwt handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833).

## Oathkeeper Removal
In one of the releases after 3.1.0, Oathkeeper will be moved to its own namespace. Support for Oathkeeper will be removed later. Once the support is removed, Oathkeeper will be installed in the clusters, but the API Gateway module will neither use it nor manage it.

### Removed Support for Oathkeeper OAuth2 Handlers
The APIRule CR in versions `v2` and `v2alpha1` does not support Oathkeeper OAuth2 handlers. Instead, it introduces the **extAuth** field, which you can use to configure an external authorizer.

**Required action**: Migrate your Oathkeeper-based OAuth2 handlers to use an external authorizer. To learn how to do this, see [SAP BTP, Kyma runtime: APIRule migration - Ory Oathkeeper-based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184) and [Configuration of the extAuth Access Strategy](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-15-api-rule-access-strategies).

### Removed Support for Oathkeeper Mutators
The APIRule CR in versions `v2` and `v2alpha1` does not support Oathkeeper mutators. Request mutators are replaced with request modifiers defined in the **spec.rule.request** section of APIRule CR. This section contains the request modification rules applied before forwarding the request to the target workload. Token mutators are not supported in APIRules `v2` and `v2alpha1`. For that, you must define your own **extAuth** configuration.

**Required action**: Migrate your rules that rely on Oathkeeper mutators to use request modifiers or an external authorizer. For more information, see [Configuration of the extAuth Access Strategy](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-15-api-rule-access-strategies) and [APIRule v2alpha1 Custom Resource](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-10-apirule-custom-resource).

## Removed Support for Opaque Tokens

The APIRule CR in versions `v2` and `v2alpha1` does not support the usage of Opaque tokens. Instead, it introduces the **extAuth** field, which you can use to configure an external authorizer.

**Required action**: Migrate your rules that use Opaque tokens to use an external authorizer. For more information, see [Configuration of the extAuth Access Strategy](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-15-api-rule-access-strategies).