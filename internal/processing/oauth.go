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

	accessStrategy := &rulev1alpha1.Authenticator{
		Handler: &rulev1alpha1.Handler{
			Name: "oauth2_introspection",
			Config: &runtime.RawExtension{
				Raw: requiredScopesJSON,
			},
		},
	}

	if oldAR != nil {
		newAR := prepareAccessRule(api, oldAR, api.Spec.Paths[0], []*rulev1alpha1.Authenticator{accessStrategy})
		err = o.updateAccessRule(ctx, newAR)
		if err != nil {
			return err
		}
	} else {
		ar := generateAccessRule(api, api.Spec.Paths[0], []*rulev1alpha1.Authenticator{accessStrategy})
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
