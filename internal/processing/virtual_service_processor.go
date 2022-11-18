package processing

import (
	"context"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VirtualServiceProcessor struct {
	Creator           VirtualServiceCreator
	Client            client.Client
	Ctx               context.Context
	OathkeeperSvc     string
	OathkeeperSvcPort uint32
	CorsConfig        *CorsConfig
	AdditionalLabels  map[string]string
	DefaultDomainName string
}

// TODO Find a better name
type VirtualServiceCreator interface {
	Create(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService
}

func (r VirtualServiceProcessor) EvaluateReconciliation(apiRule *gatewayv1beta1.APIRule) ([]*ObjectChange, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(r.Ctx, apiRule)
	if err != nil {
		return make([]*ObjectChange, 0), err
	}

	c := r.getObjectChanges(desired, actual)

	return []*ObjectChange{c}, nil
}

func (r VirtualServiceProcessor) getDesiredState(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	return r.Creator.Create(api)
}

func (r VirtualServiceProcessor) getActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	labels := GetOwnerLabels(api)

	var vsList networkingv1beta1.VirtualServiceList
	if err := r.Client.List(ctx, &vsList, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) >= 1 {
		return vsList.Items[0], nil
	} else {
		return nil, nil
	}
}

func (r VirtualServiceProcessor) getObjectChanges(desiredVs *networkingv1beta1.VirtualService, actualVs *networkingv1beta1.VirtualService) *ObjectChange {
	if actualVs != nil {
		actualVs.Spec = *desiredVs.Spec.DeepCopy()
		return NewObjectUpdateAction(actualVs)
	} else {
		return NewObjectCreateAction(desiredVs)
	}
}
