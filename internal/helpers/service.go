package helpers

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	apiv1beta1 "istio.io/api/type/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindServiceNamespace(api *gatewayv2alpha1.APIRule, rule *gatewayv2alpha1.Rule) string {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule != nil && rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace
	}
	if api != nil && api.Spec.Service != nil && api.Spec.Service.Namespace != nil {
		return *api.Spec.Service.Namespace
	}
	return api.Namespace
}

func GetLabelSelectorFromService(ctx context.Context, client client.Client, service *gatewayv2alpha1.Service, api *gatewayv2alpha1.APIRule, rule *gatewayv2alpha1.Rule) (*apiv1beta1.WorkloadSelector, error) {
	workloadSelector := apiv1beta1.WorkloadSelector{}

	if service == nil || service.Name == nil {
		return &workloadSelector, fmt.Errorf("service name is required but missing")
	}
	nsName := types.NamespacedName{Name: *service.Name}
	if service.Namespace != nil {
		nsName.Namespace = *service.Namespace
	} else {
		nsName.Namespace = FindServiceNamespace(api, rule)
	}

	if nsName.Namespace == "" {
		nsName.Namespace = "default"
	}

	svc := &corev1.Service{}
	err := client.Get(ctx, nsName, svc)
	if err != nil {
		return &workloadSelector, err
	}

	if len(svc.Spec.Selector) == 0 {
		return nil, nil
	}
	workloadSelector.MatchLabels = map[string]string{}
	for label, value := range svc.Spec.Selector {
		workloadSelector.MatchLabels[label] = value
	}
	return &workloadSelector, nil
}
