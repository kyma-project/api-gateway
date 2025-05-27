package status

import (
	"fmt"
	"strings"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

type ReconciliationV2alpha1Status struct {
	APIRuleStatus *gatewayv2alpha1.APIRuleStatus
}

func (s ReconciliationV2alpha1Status) HasError() bool {
	return s.APIRuleStatus != nil && s.APIRuleStatus.State == gatewayv2alpha1.Error
}

func (s ReconciliationV2alpha1Status) GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatus {
	if len(errorMap) == 0 {
		s.APIRuleStatus.State = gatewayv2alpha1.Ready
		s.APIRuleStatus.Description = "Reconciled successfully"
		return s
	}
	var resourceErrors []string
	for key, val := range errorMap {
		resource := func() string {
			switch key {
			case OnAPIRule:
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
	s.APIRuleStatus.State = gatewayv2alpha1.Error
	s.APIRuleStatus.Description = strings.Join(resourceErrors, "\n")
	return s
}

func (s ReconciliationV2alpha1Status) GenerateStatusFromFailures(failures []validation.Failure) ReconciliationStatus {
	if len(failures) == 0 {
		s.APIRuleStatus.State = gatewayv2alpha1.Ready
		s.APIRuleStatus.Description = "Reconciled successfully"
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
	s.APIRuleStatus.State = gatewayv2alpha1.Error
	s.APIRuleStatus.Description = "Validation errors: " + strings.Join(messages, "\n")
	return s
}

func (s ReconciliationV2alpha1Status) UpdateStatus(status any) error {
	st, ok := status.(*gatewayv2alpha1.APIRuleStatus)
	if !ok {
		return fmt.Errorf("status has unexpected type %T", status)
	}

	st.Description = s.APIRuleStatus.Description
	st.State = s.APIRuleStatus.State

	return nil
}
