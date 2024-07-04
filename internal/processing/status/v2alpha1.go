package status

import (
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/versions"
	"github.com/kyma-project/api-gateway/internal/validation"
	"strings"
)

type ReconciliationV2alpha1Status struct {
	State       gatewayv2alpha1.State
	Description string
}

// GetStatusForErrorMap combines the errors from the errorMap into a single string and sets the state to Error
// TODO: function should be extended to support conditions, when they are implemented
func (s ReconciliationV2alpha1Status) GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatus {
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

func (s ReconciliationV2alpha1Status) GenerateStatusFromFailures(failures []validation.Failure) ReconciliationStatus {
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

func (s ReconciliationV2alpha1Status) UpdateStatus() error {
	if status.ApiRuleStatusVersion() != versions.V2alpha1 {
		return fmt.Errorf("v2alpha1 status visitor cannot handle status of version %s", status.ApiRuleStatusVersion())
	}

	v2alpha1Status := status.(*gatewayv2alpha1.APIRuleStatus)
	v2alpha1Status.State = s.State
	v2alpha1Status.Description = s.Description

	return nil
}
