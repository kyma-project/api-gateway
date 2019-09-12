package processing

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type allow struct {
	vsClient          *istioClient.VirtualService
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
}

func (a *allow) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	destinationHost := ""
	var destinationPort uint32
	if a.isSecured(api.Spec.Rules[0]) {
		destinationHost = fmt.Sprintf("%s.svc.cluster.local", a.oathkeeperSvc)
		destinationPort = a.oathkeeperSvcPort
	} else {
		destinationHost = fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)
		destinationPort = *api.Spec.Service.Port
	}

	oldVS, err := a.getVirtualService(ctx, api)
	if err != nil {
		return err
	}

	if oldVS != nil {
		newVS := prepareVirtualService(api, oldVS, destinationHost, destinationPort, api.Spec.Rules[0].Path)
		return a.updateVirtualService(ctx, newVS)
	}
	vs := generateVirtualService(api, destinationHost, destinationPort, api.Spec.Rules[0].Path)
	return a.createVirtualService(ctx, vs)

}

func (a *allow) getVirtualService(ctx context.Context, api *gatewayv2alpha1.Gate) (*networkingv1alpha3.VirtualService, error) {
	vs, err := a.vsClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (a *allow) createVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return a.vsClient.Create(ctx, vs)
}

func (a *allow) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return a.vsClient.Update(ctx, vs)
}

func (a *allow) isSecured(rule gatewayv2alpha1.Rule) bool {
	if len(rule.Scopes) > 0 || len(rule.Mutators) > 0 {
		return true
	}
	return false
}
