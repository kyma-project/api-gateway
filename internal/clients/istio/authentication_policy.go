package istio

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	authenticationv1alpha1 "knative.dev/pkg/apis/istio/authentication/v1alpha1"
	crClient "sigs.k8s.io/controller-runtime/pkg/client"
)

//ForAuthenticationPolicy returns client for Istio Policy
func ForAuthenticationPolicy(crClient crClient.Client) *AuthenticationPolicy {
	return &AuthenticationPolicy{
		crClient: crClient,
	}
}

//AuthenticationPolicy .
type AuthenticationPolicy struct {
	crClient crClient.Client
}

//Create method creates Istio Policy
func (c *AuthenticationPolicy) Create(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return c.crClient.Create(ctx, ap)
}

//GetForAPI method gets Istio Policy for given Gate
func (c *AuthenticationPolicy) GetForAPI(ctx context.Context, api *gatewayv2alpha1.Gate) (*authenticationv1alpha1.Policy, error) {
	authenticationPolicyName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	return c.Get(ctx, authenticationPolicyName, api.GetNamespace())
}

//Get method gets Istio Policy for given name and namespace
func (c *AuthenticationPolicy) Get(ctx context.Context, apName, apNamespace string) (*authenticationv1alpha1.Policy, error) {
	namespacedName := crClient.ObjectKey{Namespace: apNamespace, Name: apName}
	var ap authenticationv1alpha1.Policy
	err := c.crClient.Get(ctx, namespacedName, &ap)
	if err != nil {
		return nil, err
	}
	return &ap, nil
}

//Update method updates Istio Policy
func (c *AuthenticationPolicy) Update(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return c.crClient.Update(ctx, ap)
}
