package processing

import (
	"context"
	"fmt"

	"istio.io/api/networking/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

var (
	//OwnerLabel .
	OwnerLabel = fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())
	//OwnerLabelv1alpha1 .
	OwnerLabelv1alpha1 = fmt.Sprintf("%s.%s", "apirule", gatewayv1alpha1.GroupVersion.String())
)

//Factory .
type Factory struct {
	client            client.Client
	Log               logr.Logger
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	JWKSURI           string
	corsConfig        *CorsConfig
	additionalLabels  map[string]string
	defaultDomainName string
}

//NewFactory .
func NewFactory(client client.Client, logger logr.Logger, oathkeeperSvc string, oathkeeperSvcPort uint32, jwksURI string, corsConfig *CorsConfig, additionalLabels map[string]string, defaultDomainName string) *Factory {
	return &Factory{
		client:            client,
		Log:               logger,
		oathkeeperSvc:     oathkeeperSvc,
		oathkeeperSvcPort: oathkeeperSvcPort,
		JWKSURI:           jwksURI,
		corsConfig:        corsConfig,
		additionalLabels:  additionalLabels,
		defaultDomainName: defaultDomainName,
	}
}

//CorsConfig is an internal representation of v1alpha3.CorsPolicy object
type CorsConfig struct {
	AllowOrigins []*v1beta1.StringMatch
	AllowMethods []string
	AllowHeaders []string
}

// CalculateRequiredState returns required state of all objects related to given api
func (f *Factory) CalculateRequiredState(api *gatewayv1beta1.APIRule) *State {
	var res State

	res.accessRules = make(map[string]*rulev1alpha1.Rule)
	for _, rule := range api.Spec.Rules {
		if isSecured(rule) {
			ar := generateAccessRule(api, rule, rule.AccessStrategies, f.additionalLabels, f.defaultDomainName)
			res.accessRules[ar.Spec.Match.URL] = ar
		}
	}

	//Only one vs
	vs := f.generateVirtualService(api)
	res.virtualService = vs

	return &res
}

//State represents desired or actual state of Istio Virtual Services and Oathkeeper Rules
type State struct {
	virtualService *networkingv1beta1.VirtualService
	accessRules    map[string]*rulev1alpha1.Rule
}

//GetActualState methods gets actual state of Istio Virtual Services and Oathkeeper Rules
func (f *Factory) GetActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (*State, error) {
	labels := make(map[string]string)
	labels[OwnerLabel] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	var state State

	var vsList networkingv1beta1.VirtualServiceList
	if err := f.client.List(ctx, &vsList, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) == 1 {
		state.virtualService = vsList.Items[0]
	} else {
		state.virtualService = nil
	}

	var arList rulev1alpha1.RuleList
	if err := f.client.List(ctx, &arList, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	state.accessRules = make(map[string]*rulev1alpha1.Rule)

	for i := range arList.Items {
		obj := arList.Items[i]
		state.accessRules[obj.Spec.Match.URL] = &obj
	}
	return &state, nil
}

//Patch represents diff between desired and actual state
type Patch struct {
	virtualService *objToPatch
	accessRule     map[string]*objToPatch
}

type objToPatch struct {
	action string
	obj    client.Object
}

//CalculateDiff methods compute diff between desired & actual state
func (f *Factory) CalculateDiff(requiredState *State, actualState *State) *Patch {
	arPatch := make(map[string]*objToPatch)

	for path, rule := range requiredState.accessRules {
		rulePatch := &objToPatch{}

		if actualState.accessRules[path] != nil {
			rulePatch.action = "update"
			modifyAccessRule(actualState.accessRules[path], rule)
			rulePatch.obj = actualState.accessRules[path]
		} else {
			rulePatch.action = "create"
			rulePatch.obj = rule
		}

		arPatch[path] = rulePatch
	}

	for path, rule := range actualState.accessRules {
		if requiredState.accessRules[path] == nil {
			objToDelete := &objToPatch{action: "delete", obj: rule}
			arPatch[path] = objToDelete
		}
	}

	vsPatch := &objToPatch{}
	if actualState.virtualService != nil {
		vsPatch.action = "update"
		f.updateVirtualService(actualState.virtualService, requiredState.virtualService)
		vsPatch.obj = actualState.virtualService
	} else {
		vsPatch.action = "create"
		vsPatch.obj = requiredState.virtualService
	}

	return &Patch{virtualService: vsPatch, accessRule: arPatch}
}

//ApplyDiff method applies computed diff
func (f *Factory) ApplyDiff(ctx context.Context, patch *Patch) error {

	err := f.applyObjDiff(ctx, patch.virtualService)
	if err != nil {
		return err
	}

	for _, rule := range patch.accessRule {
		err := f.applyObjDiff(ctx, rule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Factory) applyObjDiff(ctx context.Context, objToPatch *objToPatch) error {
	var err error

	switch objToPatch.action {
	case "create":
		err = f.client.Create(ctx, objToPatch.obj)
	case "update":
		err = f.client.Update(ctx, objToPatch.obj)
	case "delete":
		err = f.client.Delete(ctx, objToPatch.obj)
	}

	if err != nil {
		return err
	}

	return nil
}
