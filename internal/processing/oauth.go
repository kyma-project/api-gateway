package processing

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	accessRuleClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	internalTypes "github.com/kyma-incubator/api-gateway/internal/types/ory"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	oldVS, err := o.getVirtualService(ctx, api)
	if err != nil {
		return err
	}

	if oldVS != nil {
		newVS := prepareVirtualService(api, oldVS, o.oathkeeperSvc, o.oathkeeperSvcPort, api.Spec.Paths[0].Path)
		err = o.updateVirtualService(ctx, newVS)
		if err != nil {
			return err
		}
	} else {
		vs := generateVirtualService(api, o.oathkeeperSvc, o.oathkeeperSvcPort, api.Spec.Paths[0].Path)
		err = o.createVirtualService(ctx, vs)
		if err != nil {
			return err
		}
	}

	oldAR, err := o.getAccessRule(ctx, api)
	if err != nil {
		return err
	}

	requiredScopesJSON, err := generateRequiredScopesJSON(&api.Spec.Paths[0])
	if err != nil {
		return err
	}

	if oldAR != nil {
		newAR := o.prepareAccessRule(api, oldAR, requiredScopesJSON)
		err = o.updateAccessRule(ctx, newAR)
		if err != nil {
			return err
		}
	} else {
		ar := o.generateAccessRule(api, requiredScopesJSON)
		err = o.createAccessRule(ctx, ar)
		if err != nil {
			return err
		}
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

func (o *oauth) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return o.vsClient.Update(ctx, vs)
}

func (o *oauth) updateAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return o.arClient.Update(ctx, ar)
}

func generateOwnerRef(api *gatewayv2alpha1.Gate) k8sMeta.OwnerReference {
	return *builders.OwnerReference().
		Name(api.ObjectMeta.Name).
		APIVersion(api.TypeMeta.APIVersion).
		Kind(api.TypeMeta.Kind).
		UID(api.ObjectMeta.UID).
		Controller(true).
		Get()
}

func generateObjectMeta(api *gatewayv2alpha1.Gate) k8sMeta.ObjectMeta {
	ownerRef := generateOwnerRef(api)
	return *builders.ObjectMeta().
		Name(fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)).
		Namespace(api.ObjectMeta.Namespace).
		OwnerReference(builders.OwnerReference().From(&ownerRef)).
		Get()
}

func generateRequiredScopesJSON(path *gatewayv2alpha1.Path) ([]byte, error) {
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

func (o *oauth) generateAccessRule(api *gatewayv2alpha1.Gate, requiredScopes []byte) *rulev1alpha1.Rule {
	objectMeta := generateObjectMeta(api)

	rawConfig := &runtime.RawExtension{
		Raw: requiredScopes,
	}

	spec := &rulev1alpha1.RuleSpec{
		Upstream: &rulev1alpha1.Upstream{
			URL: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)),
		},
		Match: &rulev1alpha1.Match{
			Methods: api.Spec.Paths[0].Methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, api.Spec.Paths[0].Path),
		},
		Authorizer: &rulev1alpha1.Authorizer{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
		Mutators: api.Spec.Mutators,
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

func (o *oauth) prepareAccessRule(api *gatewayv2alpha1.Gate, ar *rulev1alpha1.Rule, requiredScopes []byte) *rulev1alpha1.Rule {
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
			Methods: api.Spec.Paths[0].Methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, api.Spec.Paths[0].Path),
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
		Mutators: api.Spec.Mutators,
	}

	ar.Spec = *spec

	return ar

}
