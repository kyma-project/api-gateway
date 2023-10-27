# Custom resources

## APIGateway custom resource <!-- {docsify-ignore} -->

The `apigateways.operator.kyma-project.io` CRD describes the kind and the format of data that APIGateway Controller uses to configure the resources.

Browse the documentation related to the APIGateway CR:
- [Specification of APIGateway CR](./apigateway/04-00-apigateway-custom-resource.md)
- [Kyma Gateway](./apigateway/04-10-kyma-gateway.md)
- [Oathkeeper dependency](./apigateway/04-20-oathkeeper.md)

## APIRule custom resource <!-- {docsify-ignore} -->

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data the APIRule Controller uses to configure the resources.

Browse the documentation related to the APIRule custom resource (CR):
- [Specification of APIRule CR](./apirule/04-10-apirule-custom-resource.md) describing all primary parameters of APIRule CR
- [Istio JWT access strategy](./apirule/04-20-apirule-istio-jwt-access-strategy.md) that explains how to configure **rules.accessStrategies** for Istio JWT
- [Comparison of Ory Oathkeeper and Istio JWT access strategies](./apirule/04-30-apirule-jwt-ory-and-istio-comparison.md)
- [APIRule Mutators](./apirule/04-40-apirule-mutators.md)
- [OAuth2 and JWT authorization](./apirule/04-50-apirule-authorizations.md)