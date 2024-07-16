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
	v2validation "github.com/kyma-project/api-gateway/internal/validation/v2alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciliation holds the components needed to reconcile an APIRule. The v2alpha1 reconciliation requires the APIRule in v2alpha1 and v1beta1 since
// not all underlying implementations have been migrated to v2alpha1 and the v1beta1 APIRule is used for those cases.
type Reconciliation struct {
	apiRuleV1beta1  *gatewayv1beta1.APIRule
	apiRuleV2alpha1 *gatewayv2alpha1.APIRule
	processors      []processing.ReconciliationProcessor
	config          processing.ReconciliationConfig
}

func (r Reconciliation) Validate(ctx context.Context, client client.Client) ([]validation.Failure, error) {

	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	var failures []validation.Failure
	apiRuleValidator := v2validation.NewAPIRuleValidator(r.apiRuleV2alpha1)
	failures = append(failures, apiRuleValidator.Validate(ctx, client, vsList)...)

	istioValidator := istioValidation.NewAPIRuleValidator(ctx, client, r.apiRuleV1beta1, r.config.DefaultDomainName)
	failures = append(failures, istioValidator.Validate(ctx, client, vsList)...)

	return failures, nil
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}

func NewReconciliation(apiRuleV2alpha1 *gatewayv2alpha1.APIRule, apiRuleV1beta1 *gatewayv1beta1.APIRule, config processing.ReconciliationConfig, log *logr.Logger) Reconciliation {
	vsProcessor := istio.Newv1beta1VirtualServiceProcessor(config, apiRuleV1beta1)
	apProcessor := istio.Newv1beta1AuthorizationPolicyProcessor(config, log, apiRuleV1beta1)
	raProcessor := istio.Newv1beta1RequestAuthenticationProcessor(config, apiRuleV1beta1)

	/*
		When implementing extauth handler, it should use the APIrule in version v2alpha1
		extAuth := NewExtAuthProcessor(config, log, apiv2alpha1)
	*/

	return Reconciliation{
		apiRuleV1beta1:  apiRuleV1beta1,
		apiRuleV2alpha1: apiRuleV2alpha1,
		processors:      []processing.ReconciliationProcessor{vsProcessor, raProcessor, apProcessor},
		config:          config,
	}
}

func findServiceNamespace(api *gatewayv2alpha1.APIRule, rule *gatewayv2alpha1.Rule) string {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule != nil && rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace
	}
	if api != nil && api.Spec.Service != nil && api.Spec.Service.Namespace != nil {
		return *api.Spec.Service.Namespace
	}

	if api != nil {
		return api.Namespace
	} else {
		return ""
	}
}
