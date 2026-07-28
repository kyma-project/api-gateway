package validation

import (
	"context"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApiRuleValidator interface {
	Validate(ctx context.Context, client client.Client, vsList networkingv1beta1.VirtualServiceList, gwList networkingv1beta1.GatewayList) []Failure
}

// Failure carries validation failures for a single attribute of an object.
type Failure struct {
	AttributePath string
	Message       string
}
