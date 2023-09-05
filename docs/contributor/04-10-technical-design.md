# Technical Design

## Kyma API Gateway Operator

The API Gateway Operator consists of two controllers that are reonciling different CRs. The decision why to use one operator with multiple controllers instead of multiple operators is described in an [Architecture Decision Record](https://github.com/kyma-project/api-gateway/issues/495).
The operator has a dependency on [Istio](https://istio.io/) and [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper), with the latter being installed by the operator itself.

The following diagram illustrates the APIRule reconciliation process and the resources created in the process:
![Kyma API Gateway Overview](../assets/operator-contributor-skr-overview.svg)

### API Gateway Controller

The API Gateway Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented with the [Kubebuilder](https://book.kubebuilder.io/) framework. 
The controller is responsible for handling the [APIGateway CR](../user/03-technical-reference/custom-resources/apigateway/01-30-apigateway-custom-resource.md).

#### Reconciliation
The [APIGateway CR](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md) is reconciled with each change. If no changes have been made, the default reconciliation interval 
of 10 hours set by the [Kubernetes controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) is used.  
In the event of a failure during the reconciliation process, the default [Kubernetes controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) exponential backoff requeue behavior is used.

### API Rule Controller

The API Rule Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented with the [Kubebuilder](https://book.kubebuilder.io/) framework. 
The controller is responsible for handling the [APIRule CR](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md).  
Additionally, the controller watches the [api-gateway-config](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md#jwt-access-strategy) to configure the JWT handler.

#### Reconciliation
The [APIRule CR](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md) is reconciled with each change. If no changes have been made, then the default reconciliation interval of one hour is used. 
This interval can be adjusted with the [API Gateway Operator paramteres](../user/03-technical-reference/configuration-parameters/01-10-api-gateway-operator-parameters.md).  
In the event of a failure during the reconciliation, the controller will perform the reconciliation again after 1 minute.

The following diagram illustrates the process of APIRule reconciliation and the resources created:

![APIRule CR Reconciliation](../assets/api-rule-reconciliation-sequence.svg)
