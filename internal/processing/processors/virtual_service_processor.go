package processors

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources/virtualservice"
)

const defaultHttpTimeout = time.Second * 180

// VirtualServiceProcessor is the generic processor that handles the Virtual Service in the reconciliation of API Rule.
type VirtualServiceProcessor struct {
	ApiRule    *gatewayv1beta1.APIRule
	Creator    VirtualServiceCreator
	Repository virtualservice.Repository
}

// VirtualServiceCreator provides the creation of a Virtual Service using the configuration in the given APIRule.
type VirtualServiceCreator interface {
	Create(api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error)
}

func (r VirtualServiceProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired, err := r.getDesiredState(r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	actual, err := r.getActualState(ctx, client, r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return []*processing.ObjectChange{changes}, nil
}

func (r VirtualServiceProcessor) getDesiredState(api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	return r.Creator.Create(api)
}

func (r VirtualServiceProcessor) getActualState(ctx context.Context, _ ctrlclient.Client, api *gatewayv1beta1.APIRule) (*networkingv1beta1.VirtualService, error) {
	vsList, err := r.Repository.GetAll(ctx, api)
	if err != nil {
		return nil, err
	}

	if len(vsList) >= 1 {
		return vsList[0], nil
	} else {
		return nil, nil
	}
}

func (r VirtualServiceProcessor) getObjectChanges(desiredVs *networkingv1beta1.VirtualService, actualVs *networkingv1beta1.VirtualService) *processing.ObjectChange {
	if actualVs != nil {
		actualVs.Spec = *desiredVs.Spec.DeepCopy()
		return processing.NewObjectUpdateAction(actualVs)
	} else {
		return processing.NewObjectCreateAction(desiredVs)
	}
}

func GetVirtualServiceHttpTimeout(apiRuleSpec gatewayv1beta1.APIRuleSpec, rule gatewayv1beta1.Rule) time.Duration {
	if rule.Timeout != nil {
		return time.Duration(*rule.Timeout) * time.Second
	}

	if apiRuleSpec.Timeout != nil {
		return time.Duration(*apiRuleSpec.Timeout) * time.Second
	}

	return defaultHttpTimeout
}
