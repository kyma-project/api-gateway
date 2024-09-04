package status

import (
	"github.com/kyma-project/api-gateway/internal/validation"
)

type ReconciliationStatus interface {
	UpdateStatus(status any) error

	GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatus
	GenerateStatusFromFailures([]validation.Failure) ReconciliationStatus

	HasError() bool
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
