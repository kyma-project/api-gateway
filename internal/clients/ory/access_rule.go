package ory

import (
	"context"
	"fmt"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	crClient "sigs.k8s.io/controller-runtime/pkg/client"
)

//ForAccessRule returns client for Ory Access Rule
func ForAccessRule(crClient crClient.Client) *AccessRule {
	return &AccessRule{
		crClient: crClient,
	}
}

//AccessRule .
type AccessRule struct {
	crClient crClient.Client
}

//Create method creates Oathkeeper Access Rule
func (c *AccessRule) Create(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return c.crClient.Create(ctx, ar)
}

//GetForAPI method gets Oathkeeper Access Rule for given APIRule
func (c *AccessRule) GetForAPI(ctx context.Context, api *gatewayv1alpha1.APIRule) (*rulev1alpha1.Rule, error) {
	accessRuleName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	return c.Get(ctx, accessRuleName, api.GetNamespace())
}

//Get method get Oathkeeper Access Rule for given name and namespace
func (c *AccessRule) Get(ctx context.Context, vsName, vsNamespace string) (*rulev1alpha1.Rule, error) {
	namespacedName := crClient.ObjectKey{Namespace: vsNamespace, Name: vsName}
	var ar rulev1alpha1.Rule
	err := c.crClient.Get(ctx, namespacedName, &ar)
	if err != nil {
		return nil, err
	}
	return &ar, nil
}

//Update method updates Oathkeeper Access Rule
func (c *AccessRule) Update(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return c.crClient.Update(ctx, ar)
}
