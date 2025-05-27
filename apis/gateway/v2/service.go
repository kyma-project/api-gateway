package v2

import (
	"context"
	"errors"
	"fmt"

	apiv1beta1 "istio.io/api/type/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindServiceNamespace(apiRule *APIRule, rule Rule) (string, error) {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace, nil
	}

	if apiRule == nil {
		return "", errors.New("apiRule is nil")
	}

	if apiRule.Spec.Service != nil && apiRule.Spec.Service.Namespace != nil {
		return *apiRule.Spec.Service.Namespace, nil
	}

	return apiRule.Namespace, nil
}

// PodSelector represents a service workload selector for a pod and the namespace of the service.
// +k8s:deepcopy-gen=false
type PodSelector struct {
	Selector  *apiv1beta1.WorkloadSelector
	Namespace string
}

func GetSelectorFromService(ctx context.Context, client client.Client, apiRule *APIRule, rule Rule) (PodSelector, error) {
	var service *Service
	if rule.Service != nil {
		service = rule.Service
	} else {
		service = apiRule.Spec.Service
	}

	if service == nil || service.Name == nil {
		return PodSelector{}, errors.New("service name is required but missing")
	}
	serviceNamespacedName := types.NamespacedName{Name: *service.Name}
	if service.Namespace != nil {
		serviceNamespacedName.Namespace = *service.Namespace
	} else {
		ns, err := FindServiceNamespace(apiRule, rule)
		if err != nil {
			return PodSelector{}, fmt.Errorf("finding service namespace: %w", err)
		}

		serviceNamespacedName.Namespace = ns
	}

	if serviceNamespacedName.Namespace == "" {
		serviceNamespacedName.Namespace = "default"
	}

	svc := &corev1.Service{}
	err := client.Get(ctx, serviceNamespacedName, svc)
	if err != nil {
		return PodSelector{}, err
	}

	if len(svc.Spec.Selector) == 0 {
		return PodSelector{}, nil
	}

	workloadSelector := apiv1beta1.WorkloadSelector{}
	workloadSelector.MatchLabels = map[string]string{}
	for label, value := range svc.Spec.Selector {
		workloadSelector.MatchLabels[label] = value
	}

	return PodSelector{
		Selector:  &workloadSelector,
		Namespace: serviceNamespacedName.Namespace,
	}, nil
}
