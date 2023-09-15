# APIRule custom resource

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data the API Gateway Controller listens for. To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

In this directory, you can find all documentation related to APIRule CR:
- [Specification of APIRule custom resource](./04-10-apirule-custom-resource.md) listing all primery parapeters of APIRule CR
- [Istio JWT access strategy](./04-20-apirule-istio-jwt-access-strategy.md) that explains how to configure **rules.accessStrategies** for Istio JWT
- [Comparizon of Ory Oathkeeper and Istio JWT access strategies](./04-30-apirule-jwt-ory-and-istio-comparison.md)
- [APIRule Mutators](./04-40-apirule-mutators.md)
- [OAuth2 and JWT authorization](./04-50-apirule-authorizations.md)