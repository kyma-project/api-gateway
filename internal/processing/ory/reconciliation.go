package ory

import (
	"context"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciliation struct {
	processors []processing.ReconciliationProcessor
	config     processing.ReconciliationConfig
}

func NewOryReconciliation(config processing.ReconciliationConfig) Reconciliation {
	acProcessor := NewAccessRuleProcessor(config)
	vsProcessor := NewVirtualServiceProcessor(config)
	apProcessor := NewAuthorizationPolicyProcessor(config)
	raProcessor := NewRequestAuthenticationProcessor(config)

	return Reconciliation{
		processors: []processing.ReconciliationProcessor{vsProcessor, raProcessor, apProcessor, acProcessor},
		config:     config,
	}
}

func (r Reconciliation) Validate(ctx context.Context, client client.Client, apiRule *gatewayv1beta1.APIRule) ([]validation.Failure, error) {
	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	validator := validation.APIRule{
		HandlerValidator:          &handlerValidator{},
		AccessStrategiesValidator: &asValidator{},
		ServiceBlockList:          r.config.ServiceBlockList,
		DomainAllowList:           r.config.DomainAllowList,
		HostBlockList:             r.config.HostBlockList,
		DefaultDomainName:         r.config.DefaultDomainName,
	}
	return validator.Validate(apiRule, vsList), nil
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}
