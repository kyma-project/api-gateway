package v2alpha1

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

func validateGateway(parentAttributePath string, gwList networkingv1beta1.GatewayList, apiRule *gatewayv2alpha1.APIRule) []validation.Failure {
	var failures []validation.Failure

	if apiRule.Spec.Gateway == nil {
		failures = append(failures, validation.Failure{
			AttributePath: parentAttributePath,
			Message:       "Gateway not specified",
		})
	} else {
		gatewayName := *apiRule.Spec.Gateway

		if !gatewayExists(gwList, gatewayName) {
			failures = append(failures, validation.Failure{
				AttributePath: parentAttributePath + ".gateway",
				Message:       "Gateway not found",
			})
		}
	}

	return failures
}

func gatewayExists(gwList networkingv1beta1.GatewayList, gatewayName string) bool {
	for _, gw := range gwList.Items {
		if fmt.Sprintf("%s/%s", gw.Namespace, gw.Name) == gatewayName {
			return true
		}
	}

	return false
}
