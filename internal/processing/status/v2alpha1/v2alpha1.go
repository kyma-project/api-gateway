package v2alpha1

import (
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/versions"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	"github.com/kyma-project/api-gateway/internal/validation"
	"strings"
)

type ReconciliationV2alpha1Status struct {
	State       gatewayv2alpha1.State
	Description string
}

// GetStatusForErrorMap combines the errors from the errorMap into a single string and sets the state to Error
// TODO: function should be extended to support conditions, when they are implemented
func (s ReconciliationV2alpha1Status) GetStatusForErrorMap(errorMap map[status.ResourceSelector][]error) status.ReconciliationStatusVisitor {
	errString := strings.Builder{}
	for selector, errors := range errorMap {
		errString.WriteString(fmt.Sprintf("%s:", selector.String()))
		for _, err := range errors {
			errString.WriteString(fmt.Sprintf(" %s", err.Error()))
		}
		errString.WriteString("\n")
	}

	s.State = gatewayv2alpha1.Error
	s.Description = errString.String()

	return s
}

func (s ReconciliationV2alpha1Status) GenerateStatusFromFailures(failures []validation.Failure) status.ReconciliationStatusVisitor {
	if len(failures) == 0 {
		return s
	}

	s.State = gatewayv2alpha1.Error
	s.Description = generateValidationDescription(failures)

	return s
}

func (s ReconciliationV2alpha1Status) HasError() bool {
	return s.State == gatewayv2alpha1.Error
}

func (s ReconciliationV2alpha1Status) VisitStatus(status status.Status) error {
	if status.ApiRuleStatusVersion() != versions.V2alpha1 {
		return fmt.Errorf("v2alpha1 status visitor cannot handle status of version %s", status.ApiRuleStatusVersion())
	}

	v2alpha1Status := status.(*gatewayv2alpha1.APIRuleStatus)
	v2alpha1Status.State = s.State
	v2alpha1Status.Description = s.Description

	return nil
}

func generateValidationDescription(failures []validation.Failure) string {
	var description string

	if len(failures) == 1 {
		description = "Validation error: "
		description += fmt.Sprintf("Attribute \"%s\": %s", failures[0].AttributePath, failures[0].Message)
	} else {
		const maxEntries = 3
		description = "Multiple validation errors: "
		for i := 0; i < len(failures) && i < maxEntries; i++ {
			description += fmt.Sprintf("\nAttribute \"%s\": %s", failures[i].AttributePath, failures[i].Message)
		}
		if len(failures) > maxEntries {
			description += fmt.Sprintf("\n%d more error(s)...", len(failures)-maxEntries)
		}
	}

	return description
}
