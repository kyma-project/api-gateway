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
)

// GetStandardLabels returns the standard label set for ExternalGateway resources
// following the same pattern as APIRule resources
func GetStandardLabels(external *externalv1alpha1.ExternalGateway) map[string]string {
	return map[string]string{
		processing.ModuleLabelKey:          processing.ApiGatewayLabelValue,
		processing.K8sManagedByLabelKey:    processing.ApiGatewayLabelValue,
		processing.K8sComponentLabelKey:    processing.ApiGatewayLabelValue,
		processing.K8sPartOfLabelKey:       processing.ApiGatewayLabelValue,
		ExternalGatewayOwnerLabelName:      external.Name,
		ExternalGatewayOwnerLabelNamespace: external.Namespace,
	}
}
