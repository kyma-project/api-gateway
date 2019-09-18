package processing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

//Factory .
type Factory struct {
	vsClient          *istioClient.VirtualService
	arClient          *oryClient.AccessRule
	Log               logr.Logger
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	JWKSURI           string
}

//NewFactory .
func NewFactory(vsClient *istioClient.VirtualService, arClient *oryClient.AccessRule, logger logr.Logger, oathkeeperSvc string, oathkeeperSvcPort uint32, jwksURI string) *Factory {
	return &Factory{
		vsClient:          vsClient,
		arClient:          arClient,
		Log:               logger,
		oathkeeperSvc:     oathkeeperSvc,
		oathkeeperSvcPort: oathkeeperSvcPort,
		JWKSURI:           jwksURI,
	}
}

// Run ?
func (f *Factory) Run(ctx context.Context, api *gatewayv1alpha1.APIRule) error {
	var destinationHost string
	var destinationPort uint32
	var err error

	for _, rule := range api.Spec.Rules {
		if isSecured(rule) {
			destinationHost = f.oathkeeperSvc
			destinationPort = f.oathkeeperSvcPort
		} else {
			destinationHost = fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)
			destinationPort = *api.Spec.Service.Port
		}
		// Create one AR per path
		err = f.processAR(ctx, api, rule.AccessStrategies)
		if err != nil {
			return err
		}
	}
	// Compile list of paths, create one VS
	err = f.processVS(ctx, api, destinationHost, destinationPort)
	if err != nil {
		return err
	}
	return nil
}

func (f *Factory) getVirtualService(ctx context.Context, api *gatewayv1alpha1.APIRule) (*networkingv1alpha3.VirtualService, error) {
	vs, err := f.vsClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (f *Factory) createVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return f.vsClient.Create(ctx, vs)
}

func (f *Factory) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return f.vsClient.Update(ctx, vs)
}

func (f *Factory) createAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return f.arClient.Create(ctx, ar)
}

func (f *Factory) updateAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return f.arClient.Update(ctx, ar)
}

func (f *Factory) getAccessRule(ctx context.Context, api *gatewayv1alpha1.APIRule) (*rulev1alpha1.Rule, error) {
	ar, err := f.arClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ar, nil
}

func (f *Factory) processVS(ctx context.Context, api *gatewayv1alpha1.APIRule, destinationHost string, destinationPort uint32) error {
	oldVS, err := f.getVirtualService(ctx, api)
	if err != nil {
		return err
	}

	if oldVS != nil {
		newVS := prepareVirtualService(api, oldVS, destinationHost, destinationPort, api.Spec.Rules[0].Path)
		return f.updateVirtualService(ctx, newVS)
	}
	vs := generateVirtualService(api, destinationHost, destinationPort, api.Spec.Rules[0].Path)
	return f.createVirtualService(ctx, vs)
}

func (f *Factory) processAR(ctx context.Context, api *gatewayv1alpha1.APIRule, config []*rulev1alpha1.Authenticator) error {
	oldAR, err := f.getAccessRule(ctx, api)
	if err != nil {
		return err
	}

	if oldAR != nil {
		ar := prepareAccessRule(api, oldAR, api.Spec.Rules[0], config)
		err = f.updateAccessRule(ctx, ar)
		if err != nil {
			return err
		}
	} else {
		ar := generateAccessRule(api, api.Spec.Rules[0], config)
		err = f.createAccessRule(ctx, ar)
		if err != nil {
			return err
		}
	}
	return nil
}
