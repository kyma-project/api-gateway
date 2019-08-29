package processing

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	authenticationv1alpha1 "knative.dev/pkg/apis/istio/authentication/v1alpha1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type jwt struct {
	vsClient *istioClient.VirtualService
	apClient *istioClient.AuthenticationPolicy
	JWKSURI  string
}

func (j *jwt) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	jwtConfig, err := j.toJWTConfig(api.Spec.Auth.Config)
	if err != nil {
		return err
	}

	switch jwtConfig.Mode.Name {
	case gatewayv2alpha1.JWTAll:
		{
			modeConfig, err := j.toJWTModeALLConfig(jwtConfig.Mode.Config)
			if err != nil {
				return err
			}
			if len(modeConfig.Scopes) == 0 {
				oldAP, err := j.getAuthenticationPolicy(ctx, api)
				if err != nil {
					return err
				}
				if oldAP != nil {
					return j.updateAuthenticationPolicy(ctx, j.prepareAuthenticationPolicy(api, jwtConfig, oldAP))
				}
				err = j.createAuthenticationPolicy(ctx, j.generateAuthenticationPolicy(api, jwtConfig))
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("scope support not yet implemented")
			}
		}
	default:
		return fmt.Errorf("unsupported mode: %s", jwtConfig.Mode.Name)
	}

	oldVS, err := j.getVirtualService(ctx, api)
	if err != nil {
		return err
	}
	if oldVS != nil {
		return j.updateVirtualService(ctx, j.prepareVirtualService(api, oldVS))
	}
	err = j.createVirtualService(ctx, j.generateVirtualService(api))
	if err != nil {
		return err
	}
	return nil
}

func (j *jwt) getVirtualService(ctx context.Context, api *gatewayv2alpha1.Gate) (*networkingv1alpha3.VirtualService, error) {
	vs, err := j.vsClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (j *jwt) createVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return j.vsClient.Create(ctx, vs)
}

func (j *jwt) prepareVirtualService(api *gatewayv2alpha1.Gate, vs *networkingv1alpha3.VirtualService) *networkingv1alpha3.VirtualService {
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

func (j *jwt) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return j.vsClient.Update(ctx, vs)
}

func (j *jwt) generateVirtualService(api *gatewayv2alpha1.Gate) *networkingv1alpha3.VirtualService {
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

func (j *jwt) getAuthenticationPolicy(ctx context.Context, api *gatewayv2alpha1.Gate) (*authenticationv1alpha1.Policy, error) {
	ap, err := j.apClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ap, nil
}

func (j *jwt) createAuthenticationPolicy(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return j.apClient.Create(ctx, ap)
}

func (j *jwt) updateAuthenticationPolicy(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return j.apClient.Update(ctx, ap)
}

func (j *jwt) prepareAuthenticationPolicy(api *gatewayv2alpha1.Gate, config *gatewayv2alpha1.JWTModeConfig, ap *authenticationv1alpha1.Policy) *authenticationv1alpha1.Policy {
	authenticationPolicyName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	controller := true

	ownerRef := &k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: &controller,
	}

	ap.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{*ownerRef}
	ap.ObjectMeta.Name = authenticationPolicyName
	ap.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	targets := []authenticationv1alpha1.TargetSelector{
		{
			Name: *api.Spec.Service.Name,
		},
	}
	peers := []authenticationv1alpha1.PeerAuthenticationMethod{
		{
			Mtls: &authenticationv1alpha1.MutualTLS{},
		},
	}
	origins := []authenticationv1alpha1.OriginAuthenticationMethod{
		{
			Jwt: &authenticationv1alpha1.Jwt{
				Issuer:  config.Issuer,
				JwksURI: j.JWKSURI,
			},
		},
	}
	spec := &authenticationv1alpha1.PolicySpec{
		Targets:          targets,
		PrincipalBinding: authenticationv1alpha1.PrincipalBindingUserOrigin,
		Peers:            peers,
		Origins:          origins,
	}
	ap.Spec = *spec
	return ap
}

func (j *jwt) generateAuthenticationPolicy(api *gatewayv2alpha1.Gate, config *gatewayv2alpha1.JWTModeConfig) *authenticationv1alpha1.Policy {
	authenticationPolicyName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	controller := true

	ownerRef := &k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: &controller,
	}

	objectMeta := k8sMeta.ObjectMeta{
		Name:            authenticationPolicyName,
		Namespace:       api.ObjectMeta.Namespace,
		OwnerReferences: []k8sMeta.OwnerReference{*ownerRef},
	}
	targets := []authenticationv1alpha1.TargetSelector{
		{
			Name: *api.Spec.Service.Name,
		},
	}
	peers := []authenticationv1alpha1.PeerAuthenticationMethod{
		{
			Mtls: &authenticationv1alpha1.MutualTLS{},
		},
	}
	origins := []authenticationv1alpha1.OriginAuthenticationMethod{
		{
			Jwt: &authenticationv1alpha1.Jwt{
				Issuer:  config.Issuer,
				JwksURI: j.JWKSURI,
			},
		},
	}
	spec := &authenticationv1alpha1.PolicySpec{
		Targets:          targets,
		PrincipalBinding: authenticationv1alpha1.PrincipalBindingUserOrigin,
		Peers:            peers,
		Origins:          origins,
	}
	ap := &authenticationv1alpha1.Policy{
		ObjectMeta: objectMeta,
		Spec:       *spec,
	}
	return ap
}

func (j *jwt) toJWTConfig(config *runtime.RawExtension) (*gatewayv2alpha1.JWTModeConfig, error) {
	var template gatewayv2alpha1.JWTModeConfig
	err := json.Unmarshal(config.Raw, &template)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &template, nil
}

func (j *jwt) toJWTModeALLConfig(config *runtime.RawExtension) (*gatewayv2alpha1.JWTModeALL, error) {
	var template gatewayv2alpha1.JWTModeALL
	err := json.Unmarshal(config.Raw, &template)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &template, nil
}
