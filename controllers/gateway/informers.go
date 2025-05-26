package gateway

import (
	"context"
	"slices"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

func NewServiceInformer(r *APIRuleReconciler) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		var apiRules gatewayv2alpha1.APIRuleList
		if err := r.Client.List(ctx, &apiRules); err != nil {
			return nil
		}

		if len(apiRules.Items) == 0 {
			return nil
		}

		var apiRulesV2alpha1 []gatewayv2alpha1.APIRule
		var apiRulesV1beta1 []gatewayv1beta1.APIRule

		for _, apiRuleV2 := range apiRules.Items {
			if apiRuleV2.Annotations == nil {
				continue
			}
			originalVersion, ok := apiRuleV2.Annotations[gatewayv2.OriginalVersionAnnotation]
			if !ok || slices.Contains([]string{"v2", "v2alpha1"}, originalVersion) {
				apiRulesV2alpha1 = append(apiRulesV2alpha1, apiRuleV2)
			}
			if originalVersion == "v1beta1" {
				converted := gatewayv1beta1.APIRule{}
				if err := converted.ConvertFrom(&apiRuleV2); err != nil {
					r.Log.Error(err, "Failed to convert APIRule v2alpha1 to v1beta1", "name", apiRuleV2.Name, "namespace", apiRuleV2.Namespace)
					continue
				}
				apiRulesV1beta1 = append(apiRulesV1beta1, converted)
			}
		}

		requests := matchAPIRulesV1WithChangedService(apiRulesV1beta1, obj)
		requests = append(requests, matchAPIRulesV2WithChangedService(apiRulesV2alpha1, obj)...)
		return requests
	})
}

func matchAPIRulesV1WithChangedService(apiRulesV1beta1 []gatewayv1beta1.APIRule, obj client.Object) []reconcile.Request {
	var requests []reconcile.Request
	for _, apiRule := range apiRulesV1beta1 {
		// match if service is exposed by an APIRule
		// and add APIRule to the reconciliation queue
		matches := func(target *gatewayv1beta1.Service) bool {
			if target == nil {
				return false
			}

			matchesNs := apiRule.Namespace == obj.GetNamespace()
			if target.Namespace != nil {
				matchesNs = *target.Namespace == obj.GetNamespace()
			}

			var matchesName bool
			if target.Name != nil {
				matchesName = *target.Name == obj.GetName()
			}

			return matchesNs && matchesName
		}
		if matches(apiRule.Spec.Service) {
			requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: apiRule.Namespace,
				Name:      apiRule.Name,
			}})
			continue
		}
		for _, rule := range apiRule.Spec.Rules {
			if matches(rule.Service) {
				requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
					Namespace: apiRule.Namespace,
					Name:      apiRule.Name,
				}})
				continue
			}
		}
	}
	return requests
}

func matchAPIRulesV2WithChangedService(apiRules []gatewayv2alpha1.APIRule, obj client.Object) []reconcile.Request {
	var requests []reconcile.Request
	for _, apiRule := range apiRules {
		// match if service is exposed by an APIRule
		// and add APIRule to the reconciliation queue
		matches := func(target *gatewayv2alpha1.Service) bool {
			if target == nil {
				return false
			}

			matchesNs := apiRule.Namespace == obj.GetNamespace()
			if target.Namespace != nil {
				matchesNs = *target.Namespace == obj.GetNamespace()
			}

			var matchesName bool
			if target.Name != nil {
				matchesName = *target.Name == obj.GetName()
			}

			return matchesNs && matchesName
		}
		if matches(apiRule.Spec.Service) {
			requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: apiRule.Namespace,
				Name:      apiRule.Name,
			}})
			continue
		}
		for _, rule := range apiRule.Spec.Rules {
			if matches(rule.Service) {
				requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
					Namespace: apiRule.Namespace,
					Name:      apiRule.Name,
				}})
				continue
			}
		}
	}
	return requests
}
