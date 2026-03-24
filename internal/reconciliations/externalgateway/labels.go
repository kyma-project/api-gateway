package externalgateway

import (
	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
)

const (
	// ExternalGatewayOwnerLabelName is the label key used to identify the ExternalGateway owner name
	ExternalGatewayOwnerLabelName = "externalgateway.gateway.kyma-project.io/name"
	// ExternalGatewayOwnerLabelNamespace is the label key used to identify the ExternalGateway owner namespace
	ExternalGatewayOwnerLabelNamespace = "externalgateway.gateway.kyma-project.io/namespace"
	// ExternalGatewayControllerName identifies this controller as the manager
	ExternalGatewayControllerName = "externalgateway-controller"
	// K8sCreatedForLabelKey is the label key used to identify what resource this was created for
	K8sCreatedForLabelKey = "app.kubernetes.io/created-for"
)

// GetStandardLabels returns the standard label set for ExternalGateway resources
// with controller-specific managed-by and created-for labels
func GetStandardLabels(external *externalv1alpha1.ExternalGateway) map[string]string {
	return map[string]string{
		processing.ModuleLabelKey:          processing.ApiGatewayLabelValue,
		processing.K8sManagedByLabelKey:    ExternalGatewayControllerName,
		processing.K8sComponentLabelKey:    processing.ApiGatewayLabelValue,
		processing.K8sPartOfLabelKey:       processing.ApiGatewayLabelValue,
		ExternalGatewayOwnerLabelName:      external.Name,
		ExternalGatewayOwnerLabelNamespace: external.Namespace,
	}
}
