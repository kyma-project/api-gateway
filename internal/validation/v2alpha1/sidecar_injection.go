package v2alpha1

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	apiv1beta1 "istio.io/api/type/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateSidecarInjection(ctx context.Context, k8sClient client.Client, parentAttributePath string, apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (problems []validation.Failure, err error) {

	podWorkloadSelector, err := getSelectorFromService(ctx, k8sClient, apiRule, rule)
	if err != nil {
		return nil, err
	}

	return validation.NewInjectionValidator(ctx, k8sClient).Validate(parentAttributePath, podWorkloadSelector.selector, podWorkloadSelector.namespace)
}

func findServiceNamespace(apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) string {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace
	}
	if apiRule != nil && apiRule.Spec.Service != nil && apiRule.Spec.Service.Namespace != nil {
		return *apiRule.Spec.Service.Namespace
	}
	return apiRule.Namespace
}

type podSelector struct {
	selector  *apiv1beta1.WorkloadSelector
	namespace string
}

func getSelectorFromService(ctx context.Context, client client.Client, apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (podSelector, error) {

	var service *gatewayv2alpha1.Service
	if rule.Service != nil {
		service = rule.Service
	} else {
		service = apiRule.Spec.Service
	}

	if service == nil || service.Name == nil {
		return podSelector{}, fmt.Errorf("service name is required but missing")
	}
	serviceNamespacedName := types.NamespacedName{Name: *service.Name}
	if service.Namespace != nil {
		serviceNamespacedName.Namespace = *service.Namespace
	} else {
		serviceNamespacedName.Namespace = findServiceNamespace(apiRule, rule)
	}

	if serviceNamespacedName.Namespace == "" {
		serviceNamespacedName.Namespace = "default"
	}

	svc := &corev1.Service{}
	err := client.Get(ctx, serviceNamespacedName, svc)
	if err != nil {
		return podSelector{}, err
	}

	if len(svc.Spec.Selector) == 0 {
		return podSelector{}, nil
	}

	workloadSelector := apiv1beta1.WorkloadSelector{}
	workloadSelector.MatchLabels = map[string]string{}
	for label, value := range svc.Spec.Selector {
		workloadSelector.MatchLabels[label] = value
	}

	return podSelector{
		selector:  &workloadSelector,
		namespace: serviceNamespacedName.Namespace,
	}, nil
}
