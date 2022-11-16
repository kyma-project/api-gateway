package processor

import (
	"context"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VirtualService struct {
	creator           VirtualServiceCreator
	client            client.Client
	ctx               context.Context
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	additionalLabels  map[string]string
	defaultDomainName string
}

func NewOryVirtualService(config processing.ReconciliationConfig) VirtualService {
	return VirtualService{
		creator: OryVirtualServiceCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
		client: config.Client,
		ctx:    config.Ctx,
	}
}

func NewIstioVirtualService(config processing.ReconciliationConfig) VirtualService {
	return VirtualService{
		creator: IstioVirtualServiceCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
		client: config.Client,
		ctx:    config.Ctx,
	}
}

// TODO Find a better name
type VirtualServiceCreator interface {
	create(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService
}

func (r VirtualService) EvaluateReconciliation(apiRule *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, gatewayv1beta1.StatusCode, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(r.ctx, apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), gatewayv1beta1.StatusSkipped, err
	}

	c := r.getObjectChanges(desired, actual)

	return []*processing.ObjectChange{c}, gatewayv1beta1.StatusOK, nil
}

func (r VirtualService) getDesiredState(api *gatewayv1beta1.APIRule) *networkingv1beta1.VirtualService {
	return r.creator.create(api)
}

func (r VirtualService) getActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	labels := processing.GetOwnerLabels(api)

	var vsList networkingv1beta1.VirtualServiceList
	if err := r.client.List(ctx, &vsList, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) >= 1 {
		return vsList.Items[0], nil
	} else {
		return nil, nil
	}
}

func (r VirtualService) getObjectChanges(desiredVs *networkingv1beta1.VirtualService, actualVs *networkingv1beta1.VirtualService) *processing.ObjectChange {
	if actualVs != nil {
		actualVs.Spec = *desiredVs.Spec.DeepCopy()
		return processing.NewObjectUpdateAction(actualVs)
	} else {
		return processing.NewObjectCreateAction(desiredVs)
	}
}
