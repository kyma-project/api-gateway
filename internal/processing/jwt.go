package processing

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	accessRuleClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	internalTypes "github.com/kyma-incubator/api-gateway/internal/types/ory"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	authenticationv1alpha1 "knative.dev/pkg/apis/istio/authentication/v1alpha1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type jwt struct {
	arClient          *accessRuleClient.AccessRule
	vsClient          *istioClient.VirtualService
	apClient          *istioClient.AuthenticationPolicy
	JWKSURI           string
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
}

var methods = []string{"GET", "POST", "PUT", "HEAD", "DELETE", "PATCH", "OPTIONS", "TRACE", "CONNECT"}

func (j *jwt) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	var destinationHost string
	var destinationPort uint32

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

				destinationHost = fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)
				destinationPort = uint32(*api.Spec.Service.Port)
			} else {
				oldAR, err := j.getAccessRule(ctx, api)
				if err != nil {
					return err
				}

				jwtConfJSON, err := generateRequiredScopesJSONForJWT(modeConfig, jwtConfig)
				if err != nil {
					return err
				}

				if oldAR != nil {
					newAR := j.prepareAccessRule(api, oldAR, jwtConfJSON)
					err = j.updateAccessRule(ctx, newAR)
					if err != nil {
						return err
					}
				} else {
					ar := j.generateAccessRule(api, jwtConfJSON)
					err = j.createAccessRule(ctx, ar)
					if err != nil {
						return err
					}
				}

				destinationHost = j.oathkeeperSvc
				destinationPort = j.oathkeeperSvcPort
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
		return j.updateVirtualService(ctx, j.prepareVirtualService(api, oldVS, destinationHost, destinationPort))
	}
	err = j.createVirtualService(ctx, j.generateVirtualService(api, destinationHost, destinationPort))
	if err != nil {
		return err
	}

	return nil
}

func (j *jwt) createAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return j.arClient.Create(ctx, ar)
}

func (j *jwt) generateAccessRule(api *gatewayv2alpha1.Gate, jwtConfig []byte) *rulev1alpha1.Rule {
	objectMeta := generateObjectMeta(api)

	rawConfig := &runtime.RawExtension{
		Raw: jwtConfig,
	}

	spec := &rulev1alpha1.RuleSpec{
		Upstream: &rulev1alpha1.Upstream{
			URL: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)),
		},
		Match: &rulev1alpha1.Match{
			Methods: methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, "/.*"),
		},
		Authorizer: &rulev1alpha1.Authorizer{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
		Authenticators: []*rulev1alpha1.Authenticator{
			{
				Handler: &rulev1alpha1.Handler{
					Name:   "jwt",
					Config: rawConfig,
				},
			},
		},
	}

	rule := &rulev1alpha1.Rule{
		ObjectMeta: objectMeta,
		Spec:       *spec,
	}

	return rule
}

func (j *jwt) updateAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return j.arClient.Update(ctx, ar)
}

func generateRequiredScopesJSONForJWT(jwtAllConf *gatewayv2alpha1.JWTModeALL, conf *gatewayv2alpha1.JWTModeConfig) ([]byte, error) {
	jwtConf := &internalTypes.JwtConfig{
		RequiredScope: jwtAllConf.Scopes,
		TrustedIssuer: []string{conf.Issuer},
	}
	return json.Marshal(jwtConf)
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

func (j *jwt) prepareVirtualService(api *gatewayv2alpha1.Gate, vs *networkingv1alpha3.VirtualService, destinationHost string, destinationPort uint32) *networkingv1alpha3.VirtualService {
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
			Host: destinationHost,
			Port: networkingv1alpha3.PortSelector{
				Number: destinationPort,
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

func (j *jwt) generateVirtualService(api *gatewayv2alpha1.Gate, destinationHost string, destinationPort uint32) *networkingv1alpha3.VirtualService {
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
			Host: destinationHost,
			Port: networkingv1alpha3.PortSelector{
				Number: destinationPort,
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

func (j *jwt) getAccessRule(ctx context.Context, api *gatewayv2alpha1.Gate) (*rulev1alpha1.Rule, error) {
	ar, err := j.arClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ar, nil
}

func (j *jwt) prepareAccessRule(api *gatewayv2alpha1.Gate, ar *rulev1alpha1.Rule, jwtConfig []byte) *rulev1alpha1.Rule {
	ar.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{generateOwnerRef(api)}
	ar.ObjectMeta.Name = fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	ar.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	rawConfig := &runtime.RawExtension{
		Raw: jwtConfig,
	}

	spec := &rulev1alpha1.RuleSpec{
		Upstream: &rulev1alpha1.Upstream{
			URL: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)),
		},
		Match: &rulev1alpha1.Match{
			Methods: methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, "/.*"),
		},
		Authorizer: &rulev1alpha1.Authorizer{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
		Authenticators: []*rulev1alpha1.Authenticator{
			{
				Handler: &rulev1alpha1.Handler{
					Name:   "jwt",
					Config: rawConfig,
				},
			},
		},
	}

	ar.Spec = *spec

	return ar

}
