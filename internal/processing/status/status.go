package status

import (
	"fmt"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

type ReconciliationStatus interface {
	UpdateStatus(status Status) error

	GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatus
	GenerateStatusFromFailures([]validation.Failure) ReconciliationStatus

	HasError() bool
}

// Status stores 2 status implementations for different APIRule types
// Serves as a compat layer
type Status struct {
	V1beta1Status  *v1beta1.APIRuleStatus
	V2alpha1Status *v2alpha1.APIRuleStatus
}

type ResourceSelector int

const (
	OnApiRule ResourceSelector = iota
	OnVirtualService
	OnAccessRule
	OnAuthorizationPolicy
	OnRequestAuthentication
)

func (r ResourceSelector) String() string {
	switch r {
	case OnVirtualService:
		return "VirtualService"
	case OnAccessRule:
		return "Rule"
	case OnRequestAuthentication:
		return "RequestAuthentication"
	case OnAuthorizationPolicy:
		return "AuthorizationPolicy"
	default:
		// If no Kind is resolved from the resource (e.g. subresource CRD is missing)
		return "APIRule"
	}
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
