package processing

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/api-gateway/internal/builders"

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

// RequiredObjects carries required state of the cluster after reconciliation
type RequiredObjects struct {
	virtualServices []*networkingv1alpha3.VirtualService
	accessRules     []*rulev1alpha1.Rule
}

// CalculateRequiredState returns required state of all objects related to given api
func (f *Factory) CalculateRequiredState(api *gatewayv1alpha1.APIRule) RequiredObjects {
	var res RequiredObjects

	for i, rule := range api.Spec.Rules {
		if isSecured(rule) {
			ar := generateAccessRule(api, api.Spec.Rules[i], i, rule.AccessStrategies)
			res.accessRules = append(res.accessRules, ar)
		}
	}

	//Only one vs
	vs := f.generateVirtualService(api)
	res.virtualServices = append(res.virtualServices, vs)

	return res
}

// ApplyRequiredState applies required state to the cluster
//TODO: It should be possible to get rid of api parameter
func (f *Factory) ApplyRequiredState(ctx context.Context, requiredState RequiredObjects, api *gatewayv1alpha1.APIRule) error {

	for i, rule := range requiredState.accessRules {
		err := f.processAR(ctx, rule, api, i)
		if err != nil {
			return err
		}
	}

	for _, vs := range requiredState.virtualServices {
		err := f.processVS(ctx, vs, api)
		if err != nil {
			return err
		}
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

func (f *Factory) getAccessRule(ctx context.Context, api *gatewayv1alpha1.APIRule, ruleInd int) (*rulev1alpha1.Rule, error) {
	ar, err := f.arClient.GetForAPI(ctx, api, ruleInd)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ar, nil
}

func (f *Factory) processVS(ctx context.Context, required *networkingv1alpha3.VirtualService, api *gatewayv1alpha1.APIRule) error {
	existingVS, err := f.getVirtualService(ctx, api)
	if err != nil {
		return err
	}
	if existingVS != nil {
		f.modifyVirtualService(existingVS, required)
		return f.updateVirtualService(ctx, existingVS)
	}
	return f.createVirtualService(ctx, required)
}

func (f *Factory) processAR(ctx context.Context, required *rulev1alpha1.Rule, api *gatewayv1alpha1.APIRule, ruleInd int) error {
	existingAR, err := f.getAccessRule(ctx, api, ruleInd)
	if err != nil {
		return err
	}

	if existingAR != nil {
		modifyAccessRule(existingAR, required)
		return f.updateAccessRule(ctx, existingAR)
	}
	return f.createAccessRule(ctx, required)
}

//TODO: Find better name (updateVirtualService is already taken)
func (f *Factory) modifyVirtualService(existing, required *networkingv1alpha3.VirtualService) {
	existing.Spec = required.Spec
}

func (f *Factory) generateVirtualService(api *gatewayv1alpha1.APIRule) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	ownerRef := generateOwnerRef(api)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.Host(*api.Spec.Service.Host)
	vsSpecBuilder.Gateway(*api.Spec.Gateway)

	for _, rule := range api.Spec.Rules {
		httpRouteBuilder := builders.HTTPRoute()

		if isSecured(rule) {
			httpRouteBuilder.Route(builders.RouteDestination().Host(f.oathkeeperSvc).Port(f.oathkeeperSvcPort))
		} else {
			destinationHost := fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)
			httpRouteBuilder.Route(builders.RouteDestination().Host(destinationHost).Port(*api.Spec.Service.Port))
		}

		httpRouteBuilder.Match(builders.MatchRequest().URI().Regex(rule.Path))
		vsSpecBuilder.HTTP(httpRouteBuilder)
	}

	vsBuilder := builders.VirtualService().
		Name(virtualServiceName).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef))

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get()
}
