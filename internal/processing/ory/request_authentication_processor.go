package ory

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// NewRequestAuthenticationProcessor returns a RequestAuthenticationProcessor with the desired state handling specific for the Istio handler.
func NewRequestAuthenticationProcessor(config processing.ReconciliationConfig) processors.RequestAuthenticationProcessor {
	return processors.RequestAuthenticationProcessor{
		Creator: requestAuthenticationCreator{
			additionalLabels: config.AdditionalLabels,
		},
	}
}

type requestAuthenticationCreator struct {
	additionalLabels map[string]string
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r requestAuthenticationCreator) Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.RequestAuthentication {
	return make(map[string]*securityv1beta1.RequestAuthentication)
}
