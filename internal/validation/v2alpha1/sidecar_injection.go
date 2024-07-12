package v2alpha1

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	apiv1beta1 "istio.io/api/type/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateSidecarInjection(ctx context.Context, k8sClient client.Client, parentAttributePath string, apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (problems []validation.Failure, err error) {

	var ruleForLabelSelector *gatewayv2alpha1.Rule
	if rule.Service != nil {
		// We only need to consider the rule's service if it's set, since rule service takes precedence over spec service
		ruleForLabelSelector = &rule
	}

	serviceWorkloadSelector, err := getLabelSelectorFromService(ctx, k8sClient, apiRule.Spec.Service, apiRule, ruleForLabelSelector)
	if err != nil {
		l, errorCtx := logr.FromContext(ctx)
		if errorCtx != nil {
			ctrl.Log.Error(errorCtx, "No logger in context.", "context", ctx)
		} else {
			l.Info("Couldn't get label selectors for service", "error", err)
		}
	}

	serviceNamespace := findServiceNamespace(apiRule, &rule)

	return validation.NewInjectionValidator(ctx, k8sClient).Validate(parentAttributePath, serviceWorkloadSelector, serviceNamespace)
}

func findServiceNamespace(api *gatewayv2alpha1.APIRule, rule *gatewayv2alpha1.Rule) string {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule != nil && rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace
	}
	if api != nil && api.Spec.Service != nil && api.Spec.Service.Namespace != nil {
		return *api.Spec.Service.Namespace
	}
	return api.Namespace
}

func getLabelSelectorFromService(ctx context.Context, client client.Client, service *gatewayv2alpha1.Service, api *gatewayv2alpha1.APIRule, rule *gatewayv2alpha1.Rule) (*apiv1beta1.WorkloadSelector, error) {
	workloadSelector := apiv1beta1.WorkloadSelector{}

	if service == nil || service.Name == nil {
		return &workloadSelector, fmt.Errorf("service name is required but missing")
	}
	nsName := types.NamespacedName{Name: *service.Name}
	if service.Namespace != nil {
		nsName.Namespace = *service.Namespace
	} else {
		nsName.Namespace = findServiceNamespace(api, rule)
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
