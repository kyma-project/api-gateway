# APIGateway custom resource

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data that APIGateway Controller uses to configure the API Gateway resources. Applying the custom resource (CR) triggers the installation of API Gateway resources, and deleting it triggers the uninstallation of those resources. To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apigateways.operator.kyma-project.io -o yaml
```

## Specification

**NOTE:** The specification of APIGateway CR has not been established yet. Once it is, you will be able to find the details in the following tables.

This table lists all the possible parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                                               | Type      | Description                                                                                                                                                                                                                                                                                                                                 |
|---------------------------------------------------------|-----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|

**Status:**

| Parameter            | Type   | Description                                                                                                                        |
|----------------------|--------|------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required) | string | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
