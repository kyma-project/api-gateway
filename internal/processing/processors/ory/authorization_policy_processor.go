package ory

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
)

// NewAuthorizationPolicyProcessor returns a AuthorizationPolicyProcessor with the desired state handling specific for the Istio handler.
func NewAuthorizationPolicyProcessor(_ processing.ReconciliationConfig, log *logr.Logger, apiRule *gatewayv1beta1.APIRule) processors.AuthorizationPolicyProcessor {
	return processors.AuthorizationPolicyProcessor{
		ApiRule: apiRule,
		Creator: authorizationPolicyCreator{},
		Log:     log,
	}
}

type authorizationPolicyCreator struct{}

// Create returns empty JwtAuthorization Policy.
func (r authorizationPolicyCreator) Create(_ context.Context, _ client.Client, _ *gatewayv1beta1.APIRule) (hashbasedstate.Desired, error) {
	return hashbasedstate.NewDesired(), nil
}
