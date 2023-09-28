# Technical Design

## Kyma API Gateway Operator

The API Gateway Operator consists of two controllers that are reconciling different CRs. To understand the reasons for using a single operator with multiple controllers instead of multiple operators, refer to the [Architecture Decision Record](https://github.com/kyma-project/api-gateway/issues/495).
The operator has a dependency on [Istio](https://istio.io/) and [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper). The operator itself installs the latter.

The following diagram illustrates the APIRule reconciliation process and the resources created in the process:
![Kyma API Gateway Overview](../assets/operator-contributor-skr-overview.svg)

### APIGateway Controller

APIGateway Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework.
The controller is responsible for handling the [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md).

#### Reconciliation
The [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md) is reconciled with each change. If no changes have been made, the reconciliation process occurs at the default interval of 10 hours,
as determined by the [Kubernetes controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime).
If there is a failure during the reconciliation process, the default behavior of the [Kubernetes controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) is to use exponential backoff requeue.

On deletion of [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md), the controller will check before deleting it if there any APIRule or Istio [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service) resources which are referring to the Kyma default [Gateway](https://istio.io/latest/docs/reference/config/networking/gateway/) (`kyma-system/kyma-gateway`). If there are any found, they will be listed in the logs of the controller and [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md)'s status will be set to `Warning` explaining that there are resources blocking the deletion. Deletion of the [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md) is protected with a finalizer named `apigateways.operator.kyma-project.io/api-gateway-reconciliation`. Once there are no more APIRule and VirtualService resources blocking the deletion, the [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md) will be deleted. Deletion of [APIGateway CR](../user/custom-resources/apigateway/04-00-apigateway-custom-resource.md) also deletes Kyma default Gateway.

### APIRule Controller

The APIRule Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework.
The controller is responsible for handling the [APIRule CR](../user/custom-resources/apirule/04-10-apirule-custom-resource.md).
Additionally, the controller watches the [`api-gateway-config`](../user/custom-resources/apirule/04-20-apirule-istio-jwt-access-strategy.md) to configure the JWT handler.

APIRule Controller has a conditional dependency to APIGateway Controller in terms of the default APIRule domain. If no domain is configured in APIGateway CR, APIRule Controller uses the default Kyma Gateway domain as the default value for creating Virtual Services.

**NOTE:** For now, you can only use the default domain in APIGateway CR. The option to configure your own domain will be added at a later time. See the [epic task](https://github.com/kyma-project/api-gateway/issues/130).

#### Reconciliation
[APIRule CR](../user/custom-resources/apirule/04-10-apirule-custom-resource.md) is reconciled with each change. If no changes have been made, process occurs at the default interval of 10 hours.
You can use the [API Gateway Operator paramteres](../user/technical-reference/05-00-api-gateway-operator-parameters.md) to adjust this interval.
In the event of a failure during the reconciliation, the controller performs the reconciliation again after one minute.

The following diagram illustrates the reconciliation process of APIRule and the created resources:

![APIRule CR Reconciliation](../assets/api-rule-reconciliation-sequence.svg)
