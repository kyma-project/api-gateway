package processing

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type passthrough struct {
	vsClient *istioClient.VirtualService
}

func (p *passthrough) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	fmt.Println("Processing API")

	destinationHost := fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)

	oldVS, err := p.getVirtualService(ctx, api)
	if err != nil {
		return err
	}

	if oldVS != nil {
		newVS := prepareVirtualService(api, oldVS, destinationHost, *api.Spec.Service.Port, "/.*")
		return p.updateVirtualService(ctx, newVS)
	}
	vs := generateVirtualService(api, destinationHost, *api.Spec.Service.Port, "/.*")
	return p.createVirtualService(ctx, vs)

}

func (p *passthrough) getVirtualService(ctx context.Context, api *gatewayv2alpha1.Gate) (*networkingv1alpha3.VirtualService, error) {
	vs, err := p.vsClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (p *passthrough) createVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return p.vsClient.Create(ctx, vs)
}

func (p *passthrough) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return p.vsClient.Update(ctx, vs)
}
