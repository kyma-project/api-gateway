# APIGateway Custom Resource

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data that APIGateway Controller uses to configure the API Gateway resources. Applying the custom resource (CR) triggers the installation of API Gateway resources, and deleting it triggers the uninstallation of those resources. The default CR has the name `default`.

To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apigateways.operator.kyma-project.io -o yaml
```

You are only allowed to have one APIGateway CR. If there are multiple APIGateway CRs in the cluster, the oldest one reconciles the module. Any additional APIGateway CR is placed in the `Error` state.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter             | Type     | Description                                                                                                             |
|-----------------------|----------|-------------------------------------------------------------------------------------------------------------------------|
| **enableKymaGateway** | **NO**   | Specifies whether the default [Kyma Gateway](./04-10-kyma-gateway.md), named `kyma-gateway`, should be created in the `kyma-system` namespace. |

**Status:**

| Parameter            | Type   | Description                                                                                                                        |
|----------------------|--------|------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required) | string | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
