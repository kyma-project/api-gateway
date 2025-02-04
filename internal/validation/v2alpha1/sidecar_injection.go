package v2alpha1

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateSidecarInjection(ctx context.Context, k8sClient client.Client, parentAttributePath string, apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (problems []validation.Failure, err error) {

	podWorkloadSelector, err := gatewayv2alpha1.GetSelectorForRule(ctx, k8sClient, apiRule, rule)
	if err != nil {
		return nil, err
	}

	return validation.NewInjectionValidator(ctx, k8sClient).Validate(parentAttributePath, podWorkloadSelector.Selector, podWorkloadSelector.Namespace)
}
