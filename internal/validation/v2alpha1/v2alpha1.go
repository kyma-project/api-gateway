package v2alpha1

import (
	"context"
	"fmt"
	"reflect"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

const (
	troubleshootingGuideURL = "https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-70-empty-apirule-spec"
)

type APIRuleValidator struct {
	ApiRule *gatewayv2alpha1.APIRule
}

func NewAPIRuleValidator(apiRule *gatewayv2alpha1.APIRule) validation.ApiRuleValidator {
	return &APIRuleValidator{
		ApiRule: apiRule,
	}
}

func (a *APIRuleValidator) Validate(
	ctx context.Context,
	client client.Client,
	vsList networkingv1beta1.VirtualServiceList,
	gwList networkingv1beta1.GatewayList,
) []validation.Failure {
	var failures []validation.Failure

	if reflect.DeepEqual(a.ApiRule.Spec, gatewayv2alpha1.APIRuleSpec{}) {
		failures = append(failures, validation.Failure{
			AttributePath: ".spec",
			Message:       fmt.Sprintf("APIRule in version v2alpha1 contains an empty spec. To troubleshoot, see %s.", troubleshootingGuideURL),
		})
	} else {
		failures = append(failures, validateRules(ctx, client, ".spec", a.ApiRule)...)
		failures = append(failures, validateHosts(".spec", vsList, gwList, a.ApiRule)...)
		failures = append(failures, validateGateway(".spec", gwList, a.ApiRule)...)
	}

	return failures
}
