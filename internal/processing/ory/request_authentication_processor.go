package ory

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/processing/processors"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// RequestAuthenticationProcessor is the generic processor that handles the Istio Request Authentications in the reconciliation of API Rule.
type RequestAuthenticationProcessor struct {
	Creator requestAuthenticationCreator
}

// RequestAuthenticationCreator provides the creation of RequestAuthentications using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type RequestAuthenticationCreator interface {
	Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.RequestAuthentication
}

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
