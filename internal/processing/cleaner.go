package processing

import (
	"context"

	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteAPIRuleSubresources(k8sClient client.Client, ctx context.Context, apiRule gatewayv1beta1.APIRule) error {
	labels := GetOwnerLabels(&apiRule)

	subresourceTypes := []client.Object{
		&securityv1beta1.AuthorizationPolicy{},
		&securityv1beta1.RequestAuthentication{},
		&networkingv1beta1.VirtualService{},
		&rulev1alpha1.Rule{},
	}

	for _, subresourceType := range subresourceTypes {
		return k8sClient.DeleteAllOf(ctx, subresourceType, client.MatchingLabels(labels))
	}

	return nil
}
