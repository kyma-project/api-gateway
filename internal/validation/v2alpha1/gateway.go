package v2alpha1

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

func validateGateway(parentAttributePath string, gwList networkingv1beta1.GatewayList, apiRule *gatewayv2alpha1.APIRule) []validation.Failure {
	var failures []validation.Failure

	return failures
}
