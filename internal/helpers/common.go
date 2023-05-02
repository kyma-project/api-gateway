package helpers

import (
	"context"
	"fmt"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	apiv1beta1 "istio.io/api/type/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindServiceNamespace(api *gatewayv1beta1.APIRule, rule *gatewayv1beta1.Rule) string {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule != nil && rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace
	}
	if api != nil && api.Spec.Service != nil && api.Spec.Service.Namespace != nil {
		return *api.Spec.Service.Namespace
	}
	return api.Namespace
}

func GetLabelSelectorFromService(ctx context.Context, client client.Client, service *gatewayv1beta1.Service) (*apiv1beta1.WorkloadSelector, error) {
	selector := apiv1beta1.WorkloadSelector{}
	svc := &corev1.Service{}
	err := client.Get(ctx, types.NamespacedName{Namespace: *service.Namespace, Name: *service.Name}, svc)
	if err != nil {
		return &selector, err
	}
	for name, value := range svc.Spec.Selector {
		selector.MatchLabels = map[string]string{name: value}
		return &selector, nil
	}
	return &selector, fmt.Errorf("no selectors defined for service %s/%s", *service.Namespace, *service.Name)
}
