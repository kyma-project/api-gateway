package ory

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"github.com/kyma-project/api-gateway/internal/subresources/requestauthentication"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewRequestAuthenticationProcessor returns a RequestAuthenticationProcessor with the desired state handling specific for the Istio handler.
func NewRequestAuthenticationProcessor(_ processing.ReconciliationConfig, apiRule *gatewayv1beta1.APIRule, client client.Client) processors.RequestAuthenticationProcessor {
	return processors.RequestAuthenticationProcessor{
		ApiRule:    apiRule,
		Creator:    requestAuthenticationCreator{},
		Repository: requestauthentication.NewRepository(client),
	}
}

type requestAuthenticationCreator struct{}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r requestAuthenticationCreator) Create(_ context.Context, _ client.Client, _ *gatewayv1beta1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	return make(map[string]*securityv1beta1.RequestAuthentication), nil
}
