# Technical Design

## Kyma API Gateway Operator

The API Gateway Operator consists of two controllers that are reconciling different CRs. To understand the reasons for using a single operator with multiple controllers instead of multiple operators, refer to the [Architecture Decision Record](https://github.com/kyma-project/api-gateway/issues/495).
The operator has a dependency on [Istio](https://istio.io/) and [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper). The operator itself installs the latter. 

The following diagram illustrates the APIRule reconciliation process and the resources created in the process:
![Kyma API Gateway Overview](../assets/operator-contributor-skr-overview.svg)

### API Gateway Controller

API Gateway Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework. 
The controller is responsible for handling the [APIGateway CR](../user/03-technical-reference/custom-resources/apigateway/01-30-apigateway-custom-resource.md).

#### Reconciliation
The [APIGateway CR](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md) is reconciled with each change. If no changes have been made, the reconciliation process occurs at the default interval of 10 hours,
as determined by the [Kubernetes controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime).  
If there is a failure during the reconciliation process, the default behavior of the [Kubernetes controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) is to use exponential backoff requeue. 

### API Rule Controller

The API Rule Controller is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), which is implemented using the [Kubebuilder](https://book.kubebuilder.io/) framework. 
The controller is responsible for handling the [APIRule CR](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md).  
Additionally, the controller watches the [`api-gateway-config`](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md#jwt-access-strategy) to configure the JWT handler.

#### Reconciliation
The [APIRule CR](../user/03-technical-reference/custom-resources/apirule/01-40-apirule-custom-resource.md) is reconciled with each change. If no changes have been made, process occurs at the default interval of 10 hours.
You can use the [API Gateway Operator paramteres](../user/03-technical-reference/configuration-parameters/01-10-api-gateway-operator-parameters.md) to adjust this interval.  
In the event of a failure during the reconciliation, the controller performs the reconciliation again after one minute.

The following diagram illustrates the reconciliation process of APIRule and the created resources:

![APIRule CR Reconciliation](../assets/api-rule-reconciliation-sequence.svg)
