# Custom Resources <!-- {docsify-ignore-all} -->

## APIGateway Custom Resource

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data that APIGateway Controller uses to configure resources.

Browse the documentation related to the APIGateway custom resource (CR):
- [Specification of APIGateway CR](./apigateway/04-00-apigateway-custom-resource.md)
- [Kyma Gateway](./apigateway/04-10-kyma-gateway.md)
- [Oathkeeper Dependency](./apigateway/04-20-oathkeeper.md)

## APIRule Custom Resource

The `apirules.gateway.kyma-project.io` CRD describes the kind and the format of data the APIRule Controller uses to configure resources. The APIRule CR is available in two versions: `v2alpha1` and `v1beta1`.

> [!WARNING]
> APIRule in version `v1beta1` will become deprecated on October 28, 2024. To prepare for the introduction of the stable APIRule in version `v2`, you can start testing the API and the migration procedure using version `v2alpha1`. APIRule `v2alpha1` was introduced for testing purposes only and is not meant for use in a production environment. For more information, see the [APIRule migration blog post](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-noauth-and-jwt-handlers/ba-p/13882833).

Browse the documentation related to the APIRule CR in version `v2alpha1`:
- [Specification of APIRule CR](./apirule/v2alpha1/04-10-apirule-custom-resource.md) describing all primary parameters of APIRule CR
- [APIRule Access Strategies](./apirule/v2alpha1/04-15-api-rule-access-strategies.md)

Browse the documentation related to the APIRule CR in version `v1beta1`:
- [Specification of APIRule CR](./apirule/04-10-apirule-custom-resource.md) describing all primary parameters of APIRule CR
- [APIRule Access Strategies](./apirule/04-15-api-rule-access-strategies.md)
- [Istio JWT Access Strategy](./apirule/04-20-apirule-istio-jwt-access-strategy.md) that explains how to configure **rules.accessStrategies** for Istio JWT
- [Comparison of Ory Oathkeeper and Istio JWT Access Strategies](./apirule/04-30-apirule-jwt-ory-and-istio-comparison.md)
- [APIRule Mutators](./apirule/04-40-apirule-mutators.md)
- [OAuth2 and JWT Authorization](./apirule/04-50-apirule-authorizations.md)
