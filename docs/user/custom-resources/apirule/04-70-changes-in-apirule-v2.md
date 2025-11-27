# Changes Introduced in APIRule v2

Learn about the changes that APIRule v2 introduces and the actions you must take to adjust your `v1beta1` resources.

See the changes introduced in the new versions:
- [A Workload Must Be in the Istio Service Mesh](#a-workload-must-be-in-the-istio-service-mesh)
- [Internal Traffic to Workloads Is Blocked by Default](#internal-traffic-to-workloads-is-blocked-by-default)
- [CORS Policy Is Not Applied by Default](#cors-policy-is-not-applied-by-default)
- [Path Specification Must Not Contain Regexp](#path-specification-must-not-contain-regexp)
- [JWT Configuration Requires Explicit Issuer URL](#jwt-configuration-requires-explicit-issuer-url)
- [Removed Support for Oathkeeper OAuth2 Handlers](#removed-support-for-oathkeeper-oauth2-handlers)
- [Removed Support for Oathkeeper Mutators](#removed-support-for-oathkeeper-mutators)
- [Removed Support for Opaque Tokens](#removed-support-for-opaque-tokens)

> [!WARNING]
> APIRule CRD `v2` is the latest stable version. Version `v1beta1` is removed in release 3.4 of the API Gateway module. 
>- All existing `v1beta1` APIRule configurations continue to function as expected.
>- APIRules `v1beta1` are no longer visible in the Kyma dashboard. You can still display them using kubectl, but the resources are displayed in the converted `v2` format.
>- You can no longer create new `v1beta1` APIRules. You also can't delete or edit the existing APIRules `v1beta1`. To make any changes, you must migrate to version `v2`.
>
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md). For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US#apirule-v1beta1-migration-timeline).

## A Workload Must Be in the Istio Service Mesh

To use APIRules in version `v2`, the workload that an APIRule exposes must be in the Istio service mesh. If the workload is not inside the Istio service mesh, the APIRule does not work as expected.

**Required action**: To add a workload to the Istio service mesh, [enable Istio sidecar proxy injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection).

## Internal Traffic to Workloads Is Blocked by Default

By default, access to the workload from internal traffic is blocked if APIRule CR in version `v2` is applied. This approach aligns with Kyma's "secure by default" principle. 
## CORS Policy Is Not Applied by Default

Version `v1beta1` applied the following CORS configuration by default:
```yaml
corsPolicy:
  allowOrigins: ["*"]
  allowMethods: ["GET", "POST", "PUT", "DELETE", "PATCH"]
  allowHeaders: ["Authorization", "Content-Type", "*"]
```

Version `v2` does not apply these default values. If the **corsPolicy** field is empty, the CORS configuration is not applied. For more information, see [architecture decision record #752](https://github.com/kyma-project/api-gateway/issues/752).

**Required action**: Configure CORS policy in the **corsPolicy** field.
> [!NOTE]
> Since not all APIs require the same level of access, adjust your CORS policy configuration according to your application's specific needs and security requirements.

If you decide to use the default CORS values defined in the APIRule `v1beta1`, you must explicitly define them in your APIRule `v2`. For preflight requests to work, you must explicitly allow the `"OPTIONS"` method in the **rules.methods** field of your APIRule.

## Path Specification Must Not Contain Regexp

APIRule in version `v2` does not support regexp in the **spec.rules.path** field of APIRule CR. Instead, it supports the use of the `{*}` and `{**}` operators. See the supported configurations:
- Use the exact path (for example, `/abc`). It matches the specified path exactly.
- Use the `{*}` operator (for example, `/foo/{*}` or `/foo/{*}/bar`).  This operator represents any request that matches the given pattern, with exactly one path segment replacing the operator.
- Use the `{**}` operator (for example, `/foo/{**}` or `/foo/{**}/bar`). This operator represents any request that matches the pattern with zero or more path segments in the operator’s place. It must be the last operator in the path.
- Use the wildcard path `/*`, which matches all paths. It’s equivalent to the exact `/{**}` path. 

When migrating a path that matches the pattern `/foo(.*)` from APIRule `v1beta1` to `v2`, you must define configurations for two separate paths: `/foo` and `/foo/{**}`.


> [!NOTE] The order of rules in the APIRule CR is important. Rules defined earlier in the list have a higher priority than those defined later. Therefore, we recommend defining rules from the most specific path to the most general.
> 
> Operators allow you to define a single APIRule that matches multiple request paths. However, this also introduces the possibility of path conflicts. A path conflict occurs when two or more APIRule resources match the same path and share at least one common HTTP method. This is why the order of rules is important.


For more information on the APIRule rules specification, see [Ordering Rules in APIRule `v2`](../apirule/04-20-significance-of-rule-path-and-method-order.md).

**Required action**: Replace regexp expressions in the **spec.rules.path** field of your APIRule CRs with the `{*}` and `{**}` operators.

## JWT Configuration Requires Explicit Issuer URL

Version `v2` of APIRule introduces an additional mandatory configuration field for JWT-based authorization - **issuer**. You must provide an explicit issuer URL in the APIRule CR. See an example configuration:

```yaml
rules:
- jwt:
    authentications:
        -   issuer: {YOUR_ISSUER_URL}
            jwksUri: {YOUR_JWKS_URI}
```
If you use Cloud Identity Services, you can find the issuer URL in the OIDC well-known configuration at `https://{YOUR_TENANT}.accounts.ondemand.com/.well-known/openid-configuration`.

**Required action**: Add the **issuer** field to your APIRule specification. For more information, see [Migrating APIRule `v1beta1` of Type **jwt** to Version `v2`](../../apirule-migration/01-83-migrate-jwt-v1beta1-to-v2.md).

### Removed Support for Oathkeeper OAuth2 Handlers
The APIRule CR in version `v2` does not support Oathkeeper OAuth2 handlers. Instead, it introduces the **extAuth** field, which you can use to configure an external authorizer.

**Required action**: Migrate your Oathkeeper-based OAuth2 handlers to use an external authorizer. To learn how to do this, see [Migrating APIRule v1beta1 of type oauth2_introspection to version v2 ](../../apirule-migration/01-84-migrate-oauth2-v1beta1-to-v2.md) and [Configuration of the extAuth Access Strategy](../apirule/04-15-api-rule-access-strategies.md#configuration-of-the-extauth-access-strategy).

### Removed Support for Oathkeeper Mutators
The APIRule CR in version `v2` does not support Oathkeeper mutators. Request mutators are replaced with request modifiers defined in the **spec.rule.request** section of the APIRule CR. This section contains the request modification rules applied before the request is forwarded to the target workload. Token mutators are not supported in APIRule `v2`. For that, you must define your own **extAuth** configuration.

**Required action**: Migrate your rules that rely on Oathkeeper mutators to use request modifiers or an external authorizer. For more information, see [Configuration of the extAuth Access Strategy](../apirule/04-15-api-rule-access-strategies.md#configuration-of-the-extauth-access-strategy).

## Removed Support for Opaque Tokens

The APIRule CR in version `v2` does not support the usage of Opaque tokens. Instead, it introduces the **extAuth** field, which you can use to configure an external authorizer.

**Required action**: Migrate your rules that use Opaque tokens to use an external authorizer. For more information, see [Configuration of the extAuth Access Strategy](../apirule/04-15-api-rule-access-strategies.md#configuration-of-the-extauth-access-strategy).