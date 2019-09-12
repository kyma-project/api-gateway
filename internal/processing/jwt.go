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

type jwt struct {
	arClient          *accessRuleClient.AccessRule
	vsClient          *istioClient.VirtualService
	JWKSURI           string
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
}

func (j *jwt) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	jwtConfig, err := j.toJWTConfig(api.Spec.Auth.Config)
	if err != nil {
		return err
	}

	oldAR, err := j.getAccessRule(ctx, api)
	if err != nil {
		return err
	}

	jwtConfJSON, err := generateRequiredScopesJSONForJWT(api, jwtConfig)
	if err != nil {
		return err
	}

	accessStrategy := &rulev1alpha1.Authenticator{
		Handler: &rulev1alpha1.Handler{
			Name: "jwt",
			Config: &runtime.RawExtension{
				Raw: jwtConfJSON,
			},
		},
	}

	if oldAR != nil {
		newAR := prepareAccessRule(api, oldAR, api.Spec.Rules[0], []*rulev1alpha1.Authenticator{accessStrategy})
		err = j.updateAccessRule(ctx, newAR)
		if err != nil {
			return err
		}
	} else {
		ar := generateAccessRule(api, api.Spec.Rules[0], []*rulev1alpha1.Authenticator{accessStrategy})
		err = j.createAccessRule(ctx, ar)
		if err != nil {
			return err
		}
	}

	oldVS, err := j.getVirtualService(ctx, api)
	if err != nil {
		return err
	}
	if oldVS != nil {
		return j.updateVirtualService(ctx, prepareVirtualService(api, oldVS, j.oathkeeperSvc, j.oathkeeperSvcPort, api.Spec.Rules[0].Path))
	}
	err = j.createVirtualService(ctx, generateVirtualService(api, j.oathkeeperSvc, j.oathkeeperSvcPort, api.Spec.Rules[0].Path))
	if err != nil {
		return err
	}

	return nil
}

func (j *jwt) createAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return j.arClient.Create(ctx, ar)
}

func (j *jwt) updateAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return j.arClient.Update(ctx, ar)
}

func generateRequiredScopesJSONForJWT(gate *gatewayv2alpha1.Gate, conf *gatewayv2alpha1.JWTModeConfig) ([]byte, error) {
	jwtConf := &internalTypes.JwtConfig{
		RequiredScope: gate.Spec.Rules[0].Scopes,
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

func prepareVirtualService(api *gatewayv2alpha1.Gate, vs *networkingv1alpha3.VirtualService, destinationHost string, destinationPort uint32, path string) *networkingv1alpha3.VirtualService {
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
					builders.MatchRequest().URI().Regex(path),
					builders.RouteDestination().Host(destinationHost).Port(destinationPort))).
		Get()
}

func (j *jwt) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return j.vsClient.Update(ctx, vs)
}

func generateVirtualService(api *gatewayv2alpha1.Gate, destinationHost string, destinationPort uint32, path string) *networkingv1alpha3.VirtualService {
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
					builders.MatchRequest().URI().Regex(path),
					builders.RouteDestination().Host(destinationHost).Port(destinationPort))).
		Get()
}

func (j *jwt) toJWTConfig(config *runtime.RawExtension) (*gatewayv2alpha1.JWTModeConfig, error) {
	var template gatewayv2alpha1.JWTModeConfig
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

func prepareAccessRule(api *gatewayv2alpha1.Gate, ar *rulev1alpha1.Rule, rule gatewayv2alpha1.Rule, accessStrategies []*rulev1alpha1.Authenticator) *rulev1alpha1.Rule {
	ar.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{generateOwnerRef(api)}
	ar.ObjectMeta.Name = fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	ar.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	spec := &rulev1alpha1.RuleSpec{
		Upstream: &rulev1alpha1.Upstream{
			URL: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)),
		},
		Match: &rulev1alpha1.Match{
			Methods: rule.Methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, rule.Path),
		},
		Authorizer: &rulev1alpha1.Authorizer{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
		Authenticators: accessStrategies,
		Mutators:       rule.Mutators,
	}

	ar.Spec = *spec

	return ar

}

func generateAccessRule(api *gatewayv2alpha1.Gate, rule gatewayv2alpha1.Rule, accessStrategies []*rulev1alpha1.Authenticator) *rulev1alpha1.Rule {
	objectMeta := generateObjectMeta(api)

	spec := &rulev1alpha1.RuleSpec{
		Upstream: &rulev1alpha1.Upstream{
			URL: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)),
		},
		Match: &rulev1alpha1.Match{
			Methods: rule.Methods,
			URL:     fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, rule.Path),
		},
		Authorizer: &rulev1alpha1.Authorizer{
			Handler: &rulev1alpha1.Handler{
				Name: "allow",
			},
		},
		Authenticators: accessStrategies,
		Mutators:       rule.Mutators,
	}

	accessRule := &rulev1alpha1.Rule{
		ObjectMeta: objectMeta,
		Spec:       *spec,
	}

	return accessRule
}
