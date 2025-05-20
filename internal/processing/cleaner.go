package processing

import (
	"context"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func DeleteAPIRuleSubresources(k8sClient client.Client, ctx context.Context, apiRule gatewayv1beta1.APIRule) error {
	labels := GetOwnerLabels(&apiRule)

	var apList securityv1beta1.AuthorizationPolicyList
	err := k8sClient.List(ctx, &apList, client.MatchingLabels(labels))
	if err != nil {
		return err
	}
	for _, ap := range apList.Items {
		log.Log.Info("Removing subresource", "AuthorizationPolicy", ap.Name)
		err := k8sClient.Delete(ctx, ap)
		if err != nil {
			return err
		}
	}

	var raList securityv1beta1.RequestAuthenticationList
	err = k8sClient.List(ctx, &raList, client.MatchingLabels(labels))
	if err != nil {
		return err
	}
	for _, ra := range raList.Items {
		log.Log.Info("Removing subresource", "RequestAuthentication", ra.Name)
		err := k8sClient.Delete(ctx, ra)
		if err != nil {
			return err
		}
	}

	var vsList networkingv1beta1.VirtualServiceList
	err = k8sClient.List(ctx, &vsList, client.MatchingLabels(labels))
	if err != nil {
		return err
	}
	for _, vs := range vsList.Items {
		log.Log.Info("Removing subresource", "VirtualService", vs.Name)
		err := k8sClient.Delete(ctx, vs)
		if err != nil {
			return err
		}
	}

	var ruleList rulev1alpha1.RuleList
	err = k8sClient.List(ctx, &ruleList, client.MatchingLabels(labels))
	if err != nil {
		return err
	}
	for _, accessRule := range ruleList.Items {
		log.Log.Info("Removing subresource", "Rule", accessRule.Name)
		err := k8sClient.Delete(ctx, &accessRule)
		if err != nil {
			return err
		}
	}

	return nil
}
