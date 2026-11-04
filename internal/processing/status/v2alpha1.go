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
	// ManagedResources marks which sub-resources are actually managed for the
	// APIRule on this pass. A per-resource condition is written only for a
	// selector present here; a sub-resource not managed gets no condition at all.
	ManagedResources map[ResourceSelector]bool
}

func (s ReconciliationV2alpha1Status) HasError() bool {
	return s.ApiRuleStatus != nil && s.ApiRuleStatus.State == gatewayv2alpha1.Error
}

func (s ReconciliationV2alpha1Status) GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatus {
	if len(errorMap) == 0 {
		s.ApiRuleStatus.State = gatewayv2alpha1.Ready
		s.ApiRuleStatus.Description = "Reconciled successfully"
		setConfiguredConditions(s.ApiRuleStatus, s.Generation, s.ManagedResources)
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
	for selector := range errorMap {
		setConditionForSelector(s.ApiRuleStatus, selector, s.Generation, s.ApiRuleStatus.Description)
	}
	setCondition(s.ApiRuleStatus, gatewayv2alpha1.ErrorCondition(s.Generation, gatewayv2alpha1.ReasonReconcileFailed, s.ApiRuleStatus.Description))
	return s
}

func (s ReconciliationV2alpha1Status) GenerateStatusFromFailures(failures []validation.Failure) ReconciliationStatus {
	if len(failures) == 0 {
		s.ApiRuleStatus.State = gatewayv2alpha1.Ready
		s.ApiRuleStatus.Description = "Reconciled successfully"
		setConfiguredConditions(s.ApiRuleStatus, s.Generation, s.ManagedResources)
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
	// Spec validation failed before any sub-resource processor ran, so only the
	// Ready error condition is written; no per-resource conditions this pass.
	setCondition(s.ApiRuleStatus, gatewayv2alpha1.ErrorCondition(s.Generation, gatewayv2alpha1.ReasonValidationFailed, s.ApiRuleStatus.Description))
	return s
}

func (s ReconciliationV2alpha1Status) UpdateStatus(status any) error {
	st, ok := status.(*gatewayv2alpha1.APIRuleStatus)
	if !ok {
		return fmt.Errorf("status has unexpected type %T", status)
	}

	st.Description = s.ApiRuleStatus.Description
	st.State = s.ApiRuleStatus.State
	for _, c := range s.ApiRuleStatus.Conditions {
		apimeta.SetStatusCondition(&st.Conditions, c)
	}

	return nil
}

func setCondition(status *gatewayv2alpha1.APIRuleStatus, condition metav1.Condition) {
	apimeta.SetStatusCondition(&status.Conditions, condition)
}

func setConfiguredConditions(status *gatewayv2alpha1.APIRuleStatus, generation int64, managed map[ResourceSelector]bool) {
	// VirtualService and AuthorizationPolicy are always managed; only RA is conditional.
	setCondition(status, gatewayv2alpha1.VirtualServiceConfiguredCondition(generation))
	if managed[OnRequestAuthentication] {
		setCondition(status, gatewayv2alpha1.RequestAuthenticationConfiguredCondition(generation))
	}
	setCondition(status, gatewayv2alpha1.AuthorizationPolicyConfiguredCondition(generation))
	setCondition(status, gatewayv2alpha1.DependenciesReadyCondition(generation))
	setCondition(status, gatewayv2alpha1.ReadyCondition(generation))
}

func setConditionForSelector(status *gatewayv2alpha1.APIRuleStatus, selector ResourceSelector, generation int64, message string) {
	switch selector {
	case OnVirtualService:
		setCondition(status, gatewayv2alpha1.VirtualServiceErrorCondition(generation, message))
	case OnRequestAuthentication:
		setCondition(status, gatewayv2alpha1.RequestAuthenticationErrorCondition(generation, message))
	case OnAuthorizationPolicy:
		setCondition(status, gatewayv2alpha1.AuthorizationPolicyErrorCondition(generation, message))
	case OnApiRule, OnAccessRule:
		// no v2 condition type; listed for exhaustive-switch linter
	}
}
