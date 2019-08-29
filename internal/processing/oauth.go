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
	"k8s.io/utils/pointer"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type oauth struct {
	arClient          *accessRuleClient.AccessRule
	vsClient          *istioClient.VirtualService
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
}

func (o *oauth) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	fmt.Println("Processing API")

	oauthConfig, err := generateOauthConfig(api)
	if err != nil {
		return err
	}

	oldVS, err := o.getVirtualService(ctx, api)
	if err != nil {
		return err
	}

	if oldVS != nil {
		newVS := o.prepareVirtualService(api, oldVS, oauthConfig)
		err = o.updateVirtualService(ctx, newVS)
		if err != nil {
			return err
		}
	} else {
		vs := o.generateVirtualService(api, oauthConfig)
		err = o.createVirtualService(ctx, vs)
		if err != nil {
			return err
		}
	}

	oldAR, err := o.getAccessRule(ctx, api)
	if err != nil {
		return err
	}

	requiredScopesJSON, err := generateRequiredScopesJSON(&oauthConfig.Paths[0])
	if err != nil {
		return err
	}

	if oldAR != nil {
		newAR := o.prepareAccessRule(api, oldAR, &oauthConfig.Paths[0], requiredScopesJSON)
		err = o.updateAccessRule(ctx, newAR)
		if err != nil {
			return err
		}
	} else {
		ar := generateAccessRule(api, &oauthConfig.Paths[0], requiredScopesJSON)
		err = o.createAccessRule(ctx, ar)
	}

	return nil
}

func (o *oauth) getVirtualService(ctx context.Context, api *gatewayv2alpha1.Gate) (*networkingv1alpha3.VirtualService, error) {
	vs, err := o.vsClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (o *oauth) getAccessRule(ctx context.Context, api *gatewayv2alpha1.Gate) (*rulev1alpha1.Rule, error) {
	ar, err := o.arClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ar, nil
}

func (o *oauth) createVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return o.vsClient.Create(ctx, vs)
}

func (o *oauth) createAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return o.arClient.Create(ctx, ar)
}

func (o *oauth) prepareVirtualService(api *gatewayv2alpha1.Gate, vs *networkingv1alpha3.VirtualService, oauthConfig *gatewayv2alpha1.OauthModeConfig) *networkingv1alpha3.VirtualService {
	vs.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{generateOwnerRef(api)}
	vs.ObjectMeta.Name = fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	vs.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	match := &networkingv1alpha3.HTTPMatchRequest{
		URI: &v1alpha1.StringMatch{
			Regex: oauthConfig.Paths[0].Path,
		},
	}
	route := &networkingv1alpha3.HTTPRouteDestination{
		Destination: networkingv1alpha3.Destination{
			Host: o.oathkeeperSvc,
			Port: networkingv1alpha3.PortSelector{
				Number: o.oathkeeperSvcPort,
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

func (o *oauth) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return o.vsClient.Update(ctx, vs)
}

func (o *oauth) updateAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return o.arClient.Update(ctx, ar)
}

func generateOwnerRef(api *gatewayv2alpha1.Gate) k8sMeta.OwnerReference {
	return k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: pointer.BoolPtr(true),
	}
}

func generateObjectMeta(api *gatewayv2alpha1.Gate) k8sMeta.ObjectMeta {
	objectMeta := k8sMeta.ObjectMeta{
		Name:            fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name),
		Namespace:       api.ObjectMeta.Namespace,
		OwnerReferences: []k8sMeta.OwnerReference{generateOwnerRef(api)},
	}

	return objectMeta
}

func (o *oauth) generateVirtualService(api *gatewayv2alpha1.Gate, oauthConfig *gatewayv2alpha1.OauthModeConfig) *networkingv1alpha3.VirtualService {
	match := &networkingv1alpha3.HTTPMatchRequest{
		URI: &v1alpha1.StringMatch{
			Regex: oauthConfig.Paths[0].Path,
		},
	}
	route := &networkingv1alpha3.HTTPRouteDestination{
		Destination: networkingv1alpha3.Destination{
			Host: o.oathkeeperSvc,
			Port: networkingv1alpha3.PortSelector{
				Number: o.oathkeeperSvcPort,
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
		ObjectMeta: generateObjectMeta(api),
		Spec:       *spec,
	}

	return vs
}

func generateRequiredScopesJSON(path *gatewayv2alpha1.Option) ([]byte, error) {
	requiredScopes := &internalTypes.OauthIntrospectionConfig{
		RequiredScope: path.Scopes}
	return json.Marshal(requiredScopes)
}

func generateOauthConfig(api *gatewayv2alpha1.Gate) (*gatewayv2alpha1.OauthModeConfig, error) {
	apiConfig := api.Spec.Auth.Config
	var oauthConfig gatewayv2alpha1.OauthModeConfig

	err := json.Unmarshal(apiConfig.Raw, &oauthConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &oauthConfig, nil
}

func generateAccessRule(api *gatewayv2alpha1.Gate, path *gatewayv2alpha1.Option, requiredScopes []byte) *rulev1alpha1.Rule {
	objectMeta := generateObjectMeta(api)

	rawConfig := &runtime.RawExtension{
		Raw: requiredScopes,
	}

	spec := &rulev1alpha1.RuleSpec{
		Upstream: &rulev1alpha1.Upstream{
			URL: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)),
		},
		Match: &rulev1alpha1.Match{
			Methods: path.Methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, path.Path),
		},
		Authorizer: &rulev1alpha1.Authorizer{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
		Authenticators: []*rulev1alpha1.Authenticator{
			{
				Handler: &rulev1alpha1.Handler{
					Name:   "oauth2_introspection",
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

func (o *oauth) prepareAccessRule(api *gatewayv2alpha1.Gate, ar *rulev1alpha1.Rule, path *gatewayv2alpha1.Option, requiredScopes []byte) *rulev1alpha1.Rule {
	ar.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{generateOwnerRef(api)}
	ar.ObjectMeta.Name = fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	ar.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	rawConfig := &runtime.RawExtension{
		Raw: requiredScopes,
	}

	spec := &rulev1alpha1.RuleSpec{
		Upstream: &rulev1alpha1.Upstream{
			URL: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)),
		},
		Match: &rulev1alpha1.Match{
			Methods: path.Methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, path.Path),
		},
		Authorizer: &rulev1alpha1.Authorizer{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
		Authenticators: []*rulev1alpha1.Authenticator{
			{
				Handler: &rulev1alpha1.Handler{
					Name:   "oauth2_introspection",
					Config: rawConfig,
				},
			},
		},
	}

	ar.Spec = *spec

	return ar

}
