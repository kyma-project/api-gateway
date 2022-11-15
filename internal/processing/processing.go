package processing

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

var (
	//OwnerLabel .
	OwnerLabel = fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())
	//OwnerLabelv1alpha1 .
	OwnerLabelv1alpha1 = fmt.Sprintf("%s.%s", "apirule", gatewayv1alpha1.GroupVersion.String())
)

// Factory .
type Factory struct {
	client      client.Client
	Log         logr.Logger
	vsProcessor VirtualServiceProcessor
	arProcessor AccessRuleProcessor
}

// NewFactory .
func NewFactory(client client.Client, logger logr.Logger, vsProcessor VirtualServiceProcessor, arProcessor AccessRuleProcessor) *Factory {
	return &Factory{
		client:      client,
		Log:         logger,
		vsProcessor: vsProcessor,
		arProcessor: arProcessor,
	}
}

// CalculateRequiredState returns required state of all objects related to given api
func (f *Factory) CalculateRequiredState(api *gatewayv1beta1.APIRule) *State {
	return &State{
		accessRules:    f.arProcessor.GetDesiredObject(api),
		virtualService: f.vsProcessor.GetDesiredObject(api),
	}
}

// GetActualState methods gets actual state of Istio Virtual Services and Oathkeeper Rules
func (f *Factory) GetActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (*State, error) {

	vs, err := f.vsProcessor.GetActualState(ctx, api)
	if err != nil {
		return nil, err
	}

	accessRules, err := f.arProcessor.GetActualState(ctx, api)
	if err != nil {
		return nil, err
	}

	return &State{
		accessRules:    accessRules,
		virtualService: vs,
	}, nil
}

// CalculateDiff methods compute diff between desired & actual state
func (f *Factory) CalculateDiff(requiredState *State, actualState *State) []*ObjToPatch {

	arsToApply := f.arProcessor.GetDiff(requiredState.accessRules, actualState.accessRules)
	vsToApply := f.vsProcessor.GetDiff(requiredState.virtualService, actualState.virtualService)
	return append(arsToApply, vsToApply)
}

// ApplyDiff method applies computed diff
func (f *Factory) ApplyDiff(ctx context.Context, objectsToApply []*ObjToPatch) error {

	for _, object := range objectsToApply {
		err := f.applyObjDiff(ctx, object)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Factory) applyObjDiff(ctx context.Context, objToPatch *ObjToPatch) error {
	var err error

	switch objToPatch.action {
	case "create":
		err = f.client.Create(ctx, objToPatch.obj)
	case "update":
		err = f.client.Update(ctx, objToPatch.obj)
	case "delete":
		err = f.client.Delete(ctx, objToPatch.obj)
	default:
		err = fmt.Errorf("apply action %s is not supported", objToPatch.action)
	}

	if err != nil {
		return err
	}

	return nil
}
