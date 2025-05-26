package processors

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"time"

	"github.com/kyma-project/api-gateway/internal/processing"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultHttpTimeout = time.Second * 180

// VirtualServiceProcessor is the generic processor that handles the Virtual Service in the reconciliation of API Rule.
type VirtualServiceProcessor struct {
	ApiRule *gatewayv2alpha1.APIRule
	Creator VirtualServiceCreator
}

// VirtualServiceCreator provides the creation of a Virtual Service using the configuration in the given APIRule.
type VirtualServiceCreator interface {
	Create(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error)
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

func (r VirtualServiceProcessor) getDesiredState(api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	return r.Creator.Create(api)
}

func (r VirtualServiceProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	labels := processing.GetOwnerLabels(api)

	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(vsList.Items) >= 1 {
		return vsList.Items[0], nil
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

func GetVirtualServiceHttpTimeout(apiRuleSpec gatewayv2alpha1.APIRuleSpec, rule gatewayv2alpha1.Rule) time.Duration {
	if rule.Timeout != nil {
		return time.Duration(*rule.Timeout) * time.Second
	}

	if apiRuleSpec.Timeout != nil {
		return time.Duration(*apiRuleSpec.Timeout) * time.Second
	}

	return defaultHttpTimeout
}
