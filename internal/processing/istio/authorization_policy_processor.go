package istio

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/processing/processors"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// AuthorizationPolicyProcessor is the generic processor that handles the Istio Authorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	Creator AuthorizationPolicyCreator
}

// AuthorizationPolicyCreator provides the creation of AuthorizationPolicies using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type AuthorizationPolicyCreator interface {
	Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.AuthorizationPolicy
}

// NewAuthorizationPolicyProcessor returns a AuthorizationPolicyProcessor with the desired state handling specific for the Istio handler.
func NewAuthorizationPolicyProcessor(config processing.ReconciliationConfig) processors.AuthorizationPolicyProcessor {
	return processors.AuthorizationPolicyProcessor{
		Creator: authorizationPolicyCreator{
			additionalLabels: config.AdditionalLabels,
		},
	}
}

type authorizationPolicyCreator struct {
	additionalLabels map[string]string
}

// Create returns the Authorization Policy using the configuration of the APIRule.
func (r authorizationPolicyCreator) Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.AuthorizationPolicy {
	pathDuplicates := processors.HasPathDuplicates(api.Spec.Rules)
	authorizationPolicies := make(map[string]*securityv1beta1.AuthorizationPolicy)
	for _, rule := range api.Spec.Rules {
		if processing.IsJwtSecured(rule) {
			ar := processors.GenerateAuthorizationPolicy(api, rule, r.additionalLabels)
			authorizationPolicies[processors.GetAuthorizationPolicyKey(pathDuplicates, ar)] = ar
		}
	}
	return authorizationPolicies
}
