package processing

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	builders "github.com/kyma-incubator/api-gateway/internal/builders"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type passthrough struct {
	vsClient *istioClient.VirtualService
}

func (p *passthrough) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	fmt.Println("Processing API")

	oldVS, err := p.getVirtualService(ctx, api)
	if err != nil {
		return err
	}

	if oldVS != nil {
		newVS := p.prepareVirtualService(api, oldVS)
		return p.updateVirtualService(ctx, newVS)
	}
	vs := p.generateVirtualService(api)
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

func (p *passthrough) prepareVirtualService(api *gatewayv2alpha1.Gate, vs *networkingv1alpha3.VirtualService) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)

	ownerRef := generateOwnerRef(api)
	return builders.VirtualService().From(vs).
		Name(virtualServiceName).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(
			builders.VirtualServiceSpec().
				Host(*api.Spec.Service.Host).
				Gateway(*api.Spec.Gateway).
				HTTP(
					builders.MatchRequest().URI().Regex("/.*"),
					builders.RouteDestination().
						Host(fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)).
						Port(*api.Spec.Service.Port))).
		Get()
}

func (p *passthrough) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return p.vsClient.Update(ctx, vs)
}

func (p *passthrough) generateVirtualService(api *gatewayv2alpha1.Gate) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)

	ownerRef := generateOwnerRef(api)
	return builders.VirtualService().
		Name(virtualServiceName).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(
			builders.VirtualServiceSpec().
				Host(*api.Spec.Service.Host).
				Gateway(*api.Spec.Gateway).
				HTTP(
					builders.MatchRequest().URI().Regex("/.*"),
					builders.RouteDestination().
						Host(fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)).
						Port(*api.Spec.Service.Port))).
		Get()
}
