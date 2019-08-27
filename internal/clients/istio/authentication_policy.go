package istio

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	authenticationv1alpha1 "knative.dev/pkg/apis/istio/authentication/v1alpha1"
	crClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func ForAuthenticationPolicy(crClient crClient.Client) *AuthenticationPolicy {
	return &AuthenticationPolicy{
		crClient: crClient,
	}
}

type AuthenticationPolicy struct {
	crClient crClient.Client
}

func (c *AuthenticationPolicy) Create(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return c.crClient.Create(ctx, ap)
}

func (c *AuthenticationPolicy) GetForAPI(ctx context.Context, api *gatewayv2alpha1.Gate) (*authenticationv1alpha1.Policy, error) {
	authenticationPolicyName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	return c.Get(ctx, authenticationPolicyName, api.GetNamespace())
}

func (c *AuthenticationPolicy) Get(ctx context.Context, apName, apNamespace string) (*authenticationv1alpha1.Policy, error) {
	namespacedName := crClient.ObjectKey{Namespace: apNamespace, Name: apName}
	var ap authenticationv1alpha1.Policy
	err := c.crClient.Get(ctx, namespacedName, &ap)
	if err != nil {
		return nil, err
	}
	return &ap, nil
}

func (c *AuthenticationPolicy) Update(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return c.crClient.Update(ctx, ap)
}
