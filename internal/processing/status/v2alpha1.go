package status

import (
	"fmt"
	"strings"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReconciliationV2alpha1Status struct {
	ApiRuleStatus *gatewayv2alpha1.APIRuleStatus
	Generation    int64
}

func (s ReconciliationV2alpha1Status) HasError() bool {
	return s.ApiRuleStatus != nil && s.ApiRuleStatus.State == gatewayv2alpha1.Error
}

func (s ReconciliationV2alpha1Status) GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatus {
	if len(errorMap) == 0 {
		s.ApiRuleStatus.State = gatewayv2alpha1.Ready
		s.ApiRuleStatus.Description = "Reconciled successfully"
		setConfiguredConditions(s.ApiRuleStatus, s.Generation)
		return s
	}
	var resourceErrors []string
	for key, val := range errorMap {
		resource := func() string {
			switch key {
			case OnApiRule:
				return "ApiRuleErrors"
			case OnVirtualService:
				return "VirtualServiceErrors"
			case OnRequestAuthentication:
				return "RequestAuthenticationErrors"
			case OnAuthorizationPolicy:
				return "AuthorizationPolicyErrors"
			case OnAccessRule:
				return "AccessRuleErrors"
			}
			return "OtherErrors"
		}
		var errs []string
		for _, err := range val {
			errs = append(errs, err.Error())
		}

		resourceErrors = append(resourceErrors, fmt.Sprintf("%s: %s", resource(), strings.Join(errs, ", ")))
	}
	s.ApiRuleStatus.State = gatewayv2alpha1.Error
	s.ApiRuleStatus.Description = strings.Join(resourceErrors, "\n")
	// Seed Unknown first so sub-resources skipped by the fail-fast pipeline stay
	// Unknown; actual failures are merged on top.
	setProcessingConditions(s.ApiRuleStatus, s.Generation)
	for selector := range errorMap {
		setConditionForSelector(s.ApiRuleStatus, selector, s.Generation, s.ApiRuleStatus.Description)
	}
	setCondition(s.ApiRuleStatus, gatewayv2alpha1.SubresourcesErrorCondition(s.Generation, s.ApiRuleStatus.Description))
	setCondition(s.ApiRuleStatus, gatewayv2alpha1.ErrorConditionWithGeneration(s.Generation, reasonForErrorMap(errorMap), s.ApiRuleStatus.Description))
	return s
}

func (s ReconciliationV2alpha1Status) GenerateStatusFromFailures(failures []validation.Failure) ReconciliationStatus {
	if len(failures) == 0 {
		s.ApiRuleStatus.State = gatewayv2alpha1.Ready
		s.ApiRuleStatus.Description = "Reconciled successfully"
		setConfiguredConditions(s.ApiRuleStatus, s.Generation)
		return s
	}

	var messages []string
	const maxEntries = 3
	for i := 0; i < len(failures) && i < maxEntries; i++ {
		messages = append(messages, fmt.Sprintf("Attribute '%s': %s", failures[i].AttributePath, failures[i].Message))
	}
	if len(failures) > maxEntries {
		messages = append(messages, fmt.Sprintf("%d more error(s)...", len(failures)-maxEntries))
	}
	s.ApiRuleStatus.State = gatewayv2alpha1.Error
	s.ApiRuleStatus.Description = "Validation errors: " + strings.Join(messages, "\n")
	// No sub-resource processor runs when spec validation fails; seed Unknown.
	setProcessingConditions(s.ApiRuleStatus, s.Generation)
	setCondition(s.ApiRuleStatus, gatewayv2alpha1.SubresourcesErrorCondition(s.Generation, s.ApiRuleStatus.Description))
	setCondition(s.ApiRuleStatus, gatewayv2alpha1.ErrorConditionWithGeneration(s.Generation, gatewayv2alpha1.ReasonValidationFailed, s.ApiRuleStatus.Description))
	return s
}

func (s ReconciliationV2alpha1Status) UpdateStatus(status any) error {
	st, ok := status.(*gatewayv2alpha1.APIRuleStatus)
	if !ok {
		return fmt.Errorf("status has unexpected type %T", status)
	}

	st.Description = s.ApiRuleStatus.Description
	st.State = s.ApiRuleStatus.State
	st.Conditions = s.ApiRuleStatus.Conditions

	return nil
}

func setCondition(status *gatewayv2alpha1.APIRuleStatus, condition metav1.Condition) {
	apimeta.SetStatusCondition(&status.Conditions, condition)
}

func reasonForResourceSelector(selector ResourceSelector) string {
	switch selector {
	case OnVirtualService:
		return gatewayv2alpha1.ReasonVirtualServiceReconcileFailed
	case OnRequestAuthentication:
		return gatewayv2alpha1.ReasonRequestAuthenticationReconcileFailed
	case OnAuthorizationPolicy:
		return gatewayv2alpha1.ReasonAuthorizationPolicyReconcileFailed
	default:
		return gatewayv2alpha1.ReasonReconcileFailed
	}
}

func setConfiguredConditions(status *gatewayv2alpha1.APIRuleStatus, generation int64) {
	setCondition(status, gatewayv2alpha1.VirtualServiceConfiguredCondition(generation))
	setCondition(status, gatewayv2alpha1.RequestAuthenticationConfiguredCondition(generation))
	setCondition(status, gatewayv2alpha1.AuthorizationPolicyConfiguredCondition(generation))
	setCondition(status, gatewayv2alpha1.SubresourcesConfiguredCondition(generation))
	setCondition(status, gatewayv2alpha1.ReadyConditionWithGeneration(generation))
}

// setProcessingConditions seeds every per-resource condition (and Ready) to
// Unknown, so conditions merged on top leave skipped sub-resources at Unknown.
func setProcessingConditions(status *gatewayv2alpha1.APIRuleStatus, generation int64) {
	for _, c := range gatewayv2alpha1.ProcessingConditions(generation) {
		setCondition(status, c)
	}
}

func setConditionForSelector(status *gatewayv2alpha1.APIRuleStatus, selector ResourceSelector, generation int64, message string) {
	switch selector {
	case OnVirtualService:
		setCondition(status, gatewayv2alpha1.VirtualServiceErrorCondition(generation, message))
	case OnRequestAuthentication:
		setCondition(status, gatewayv2alpha1.RequestAuthenticationErrorCondition(generation, message))
	case OnAuthorizationPolicy:
		setCondition(status, gatewayv2alpha1.AuthorizationPolicyErrorCondition(generation, message))
	}
}

func reasonForErrorMap(errorMap map[ResourceSelector][]error) string {
	if len(errorMap) != 1 {
		return gatewayv2alpha1.ReasonReconcileFailed
	}

	for selector := range errorMap {
		return reasonForResourceSelector(selector)
	}

	return gatewayv2alpha1.ReasonReconcileFailed
}
