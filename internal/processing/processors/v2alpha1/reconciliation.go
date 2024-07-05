package v2alpha1

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	"github.com/kyma-project/api-gateway/internal/validation"
	istioValidation "github.com/kyma-project/api-gateway/internal/validation/v1beta1/istio"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciliation struct {
	apiRulev1beta1  *gatewayv1beta1.APIRule
	apiRulev2alpha1 *gatewayv2alpha1.APIRule
	processors      []processing.ReconciliationProcessor
	config          processing.ReconciliationConfig
}

func (r Reconciliation) Validate(ctx context.Context, client client.Client) ([]validation.Failure, error) {

	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	// TODO: Use v2alpha1 validation
	validator := istioValidation.NewAPIRuleValidator(ctx, client, r.apiRulev1beta1, r.config.DefaultDomainName)
	return validator.Validate(ctx, client, vsList), nil
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}

func NewReconciliation(apiv2alpha1 *gatewayv2alpha1.APIRule, apiv1beta1 *gatewayv1beta1.APIRule, config processing.ReconciliationConfig, log *logr.Logger) Reconciliation {
	//TODO: Switch implementation to v2alpha1
	vsProcessor := istio.Newv1beta1VirtualServiceProcessor(config, apiv1beta1)
	apProcessor := istio.Newv1beta1AuthorizationPolicyProcessor(config, log, apiv1beta1)
	raProcessor := istio.Newv1beta1RequestAuthenticationProcessor(config, apiv1beta1)
	// End todo

	/*
		When implementing extauth handler, it should use the APIrule in version v2alpha1
		extAuth := NewExtAuthProcessor(config, log, apiv2alpha1)
	*/

	return Reconciliation{
		apiRulev1beta1:  apiv1beta1,
		apiRulev2alpha1: apiv2alpha1,
		processors:      []processing.ReconciliationProcessor{vsProcessor, raProcessor, apProcessor},
		config:          config,
	}
}
