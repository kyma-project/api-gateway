package istio

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciliation struct {
	processors []processing.ReconciliationProcessor
	config     processing.ReconciliationConfig
}

func NewIstioReconciliation(config processing.ReconciliationConfig, log *logr.Logger) Reconciliation {
	acProcessor := NewAccessRuleProcessor(config)
	vsProcessor := NewVirtualServiceProcessor(config)
	apProcessor := NewAuthorizationPolicyProcessor(config, log)
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

	validator := validation.APIRuleValidator{
		HandlerValidator:          &handlerValidator{},
		AccessStrategiesValidator: &asValidator{},
		MutatorsValidator:         &mutatorsValidator{},
		InjectionValidator:        &injectionValidator{ctx: ctx, client: client},
		RulesValidator:            &rulesValidator{},
		DefaultDomainName:         r.config.DefaultDomainName,
	}
	return validator.Validate(ctx, client, apiRule, vsList), nil
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}
