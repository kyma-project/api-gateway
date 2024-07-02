package status

import (
	"github.com/kyma-project/api-gateway/apis/gateway/versions"
	"github.com/kyma-project/api-gateway/internal/validation"
)

type ReconciliationStatusVisitor interface {
	VisitStatus(status Status) error

	GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatusVisitor
	GenerateStatusFromFailures([]validation.Failure) ReconciliationStatusVisitor

	HasError() bool
}

type Status interface {
	ApiRuleStatusVersion() versions.Version
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
