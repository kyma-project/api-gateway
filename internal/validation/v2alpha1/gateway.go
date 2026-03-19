package v2alpha1

import (
	"fmt"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

func validateGateway(parentAttributePath string, gwList networkingv1beta1.GatewayList, externalGwList externalv1alpha1.ExternalGatewayList, apiRule *gatewayv2alpha1.APIRule) []validation.Failure {
	var failures []validation.Failure

	hasGateway := apiRule.Spec.Gateway != nil
	hasExternalGateway := apiRule.Spec.ExternalGateway != nil

	// Exactly one of Gateway or ExternalGateway must be specified
	if !hasGateway && !hasExternalGateway {
		failures = append(failures, validation.Failure{
			AttributePath: parentAttributePath,
			Message:       "Either gateway or externalGateway must be specified",
		})
	} else if hasGateway && hasExternalGateway {
		failures = append(failures, validation.Failure{
			AttributePath: parentAttributePath,
			Message:       "Only one of gateway or externalGateway can be specified",
		})
	} else if hasGateway {
		gatewayName := *apiRule.Spec.Gateway
		if !gatewayExists(gwList, gatewayName) {
			failures = append(failures, validation.Failure{
				AttributePath: parentAttributePath + ".gateway",
				Message:       "Gateway not found",
			})
		}
	} else if hasExternalGateway {
		externalGatewayName := *apiRule.Spec.ExternalGateway
		if !externalGatewayExists(externalGwList, externalGatewayName) {
			failures = append(failures, validation.Failure{
				AttributePath: parentAttributePath + ".externalGateway",
				Message:       "ExternalGateway not found",
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

func externalGatewayExists(externalGwList externalv1alpha1.ExternalGatewayList, externalGatewayName string) bool {
	for _, externalGw := range externalGwList.Items {
		if fmt.Sprintf("%s/%s", externalGw.Namespace, externalGw.Name) == externalGatewayName {
			return true
		}
	}

	return false
}
