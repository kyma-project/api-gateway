package ory

import (
	"context"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewAuthorizationPolicyProcessor returns a AuthorizationPolicyProcessor with the desired state handling specific for the Istio handler.
func NewAuthorizationPolicyProcessor(config processing.ReconciliationConfig, log *logr.Logger) processors.AuthorizationPolicyProcessor {
	return processors.AuthorizationPolicyProcessor{
		Creator: authorizationPolicyCreator{
			additionalLabels: config.AdditionalLabels,
		},
		Log: log,
	}
}

type authorizationPolicyCreator struct {
	additionalLabels map[string]string
}

// Create returns empty JwtAuthorization Policy
func (r authorizationPolicyCreator) Create(ctx context.Context, client client.Client, _ *gatewayv1beta1.APIRule) (hashbasedstate.Desired, error) {
	return hashbasedstate.NewDesired(), nil
}
