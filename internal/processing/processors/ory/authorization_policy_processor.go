package ory

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"github.com/kyma-project/api-gateway/internal/subresources/authorizationpolicy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewAuthorizationPolicyProcessor returns a AuthorizationPolicyProcessor with the desired state handling specific for the Istio handler.
func NewAuthorizationPolicyProcessor(_ processing.ReconciliationConfig, log *logr.Logger, apiRule *gatewayv1beta1.APIRule, client client.Client) processors.AuthorizationPolicyProcessor {
	return processors.AuthorizationPolicyProcessor{
		ApiRule:    apiRule,
		Creator:    authorizationPolicyCreator{},
		Log:        log,
		Repository: authorizationpolicy.NewRepository(client),
	}
}

type authorizationPolicyCreator struct{}

// Create returns empty JwtAuthorization Policy
func (r authorizationPolicyCreator) Create(_ context.Context, _ client.Client, _ *gatewayv1beta1.APIRule) (hashbasedstate.Desired, error) {
	return hashbasedstate.NewDesired(), nil
}
