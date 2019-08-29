package processing

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
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
	controller := true

	ownerRef := &k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: &controller,
	}

	vs.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{*ownerRef}
	vs.ObjectMeta.Name = virtualServiceName
	vs.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	match := &networkingv1alpha3.HTTPMatchRequest{
		URI: &v1alpha1.StringMatch{
			Regex: "/.*",
		},
	}
	route := &networkingv1alpha3.HTTPRouteDestination{
		Destination: networkingv1alpha3.Destination{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace),
			Port: networkingv1alpha3.PortSelector{
				Number: uint32(*api.Spec.Service.Port),
			},
		},
	}

	spec := &networkingv1alpha3.VirtualServiceSpec{
		Hosts:    []string{*api.Spec.Service.Host},
		Gateways: []string{*api.Spec.Gateway},
		HTTP: []networkingv1alpha3.HTTPRoute{
			{
				Match: []networkingv1alpha3.HTTPMatchRequest{*match},
				Route: []networkingv1alpha3.HTTPRouteDestination{*route},
			},
		},
	}

	vs.Spec = *spec

	return vs

}

func (p *passthrough) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return p.vsClient.Update(ctx, vs)
}

func (p *passthrough) generateVirtualService(api *gatewayv2alpha1.Gate) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	controller := true

	ownerRef := &k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: &controller,
	}

	objectMeta := k8sMeta.ObjectMeta{
		Name:            virtualServiceName,
		Namespace:       api.ObjectMeta.Namespace,
		OwnerReferences: []k8sMeta.OwnerReference{*ownerRef},
	}

	match := &networkingv1alpha3.HTTPMatchRequest{
		URI: &v1alpha1.StringMatch{
			Regex: "/.*",
		},
	}
	route := &networkingv1alpha3.HTTPRouteDestination{
		Destination: networkingv1alpha3.Destination{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace),
			Port: networkingv1alpha3.PortSelector{
				Number: uint32(*api.Spec.Service.Port),
			},
		},
	}

	spec := &networkingv1alpha3.VirtualServiceSpec{
		Hosts:    []string{*api.Spec.Service.Host},
		Gateways: []string{*api.Spec.Gateway},
		HTTP: []networkingv1alpha3.HTTPRoute{
			{
				Match: []networkingv1alpha3.HTTPMatchRequest{*match},
				Route: []networkingv1alpha3.HTTPRouteDestination{*route},
			},
		},
	}

	vs := &networkingv1alpha3.VirtualService{
		ObjectMeta: objectMeta,
		Spec:       *spec,
	}

	return vs
}
