package v2alpha1

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIRuleValidator struct {
	ApiRule *gatewayv2alpha1.APIRule
}

func NewAPIRuleValidator(apiRule *gatewayv2alpha1.APIRule) *APIRuleValidator {
	return &APIRuleValidator{
		ApiRule: apiRule,
	}
}

func (a *APIRuleValidator) Validate(ctx context.Context, client client.Client, vsList networkingv1beta1.VirtualServiceList) []validation.Failure {
	var failures []validation.Failure

	failures = append(failures, validateRules(ctx, client, ".spec.rules", a.ApiRule)...)

	validation.NewInjectionValidator(ctx, client)
	return failures
}
