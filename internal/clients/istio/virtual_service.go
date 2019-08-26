package istio

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	crClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func ForVirtualService(crClient crClient.Client) *VirtualService {
	return &VirtualService{
		crClient: crClient,
	}
}

type VirtualService struct {
	crClient crClient.Client
}

func (c *VirtualService) Create(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return c.crClient.Create(ctx, vs)
}

func (c *VirtualService) GetForAPI(ctx context.Context, api *gatewayv2alpha1.Gate) (*networkingv1alpha3.VirtualService, error) {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	return c.Get(ctx, virtualServiceName, api.GetNamespace())
}

func (c *VirtualService) Get(ctx context.Context, vsName, vsNamespace string) (*networkingv1alpha3.VirtualService, error) {
	namespacedName := crClient.ObjectKey{Namespace: vsNamespace, Name: vsName}
	var vs networkingv1alpha3.VirtualService
	err := c.crClient.Get(ctx, namespacedName, &vs)
	if err != nil {
		return nil, err
	}
	return &vs, nil
}

func (c *VirtualService) Update(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return c.crClient.Update(ctx, vs)
}
