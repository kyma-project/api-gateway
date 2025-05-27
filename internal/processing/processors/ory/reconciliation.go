package ory

import (
	"context"

	"github.com/go-logr/logr"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/validation"
	oryValidation "github.com/kyma-project/api-gateway/internal/validation/v1beta1/ory"
)

type Reconciliation struct {
	apiRule    *gatewayv1beta1.APIRule
	processors []processing.ReconciliationProcessor
	config     processing.ReconciliationConfig
}

func NewOryReconciliation(apiRule *gatewayv1beta1.APIRule, config processing.ReconciliationConfig, log *logr.Logger) Reconciliation {
	acProcessor := NewAccessRuleProcessor(config, apiRule)
	vsProcessor := NewVirtualServiceProcessor(config, apiRule)
	apProcessor := NewAuthorizationPolicyProcessor(config, log, apiRule)
	raProcessor := NewRequestAuthenticationProcessor(config, apiRule)

	return Reconciliation{
		apiRule:    apiRule,
		processors: []processing.ReconciliationProcessor{vsProcessor, raProcessor, apProcessor, acProcessor},
		config:     config,
	}
}

func (r Reconciliation) Validate(ctx context.Context, client client.Client) ([]validation.Failure, error) {
	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	var gwList networkingv1beta1.GatewayList
	if err := client.List(ctx, &gwList); err != nil {
		return make([]validation.Failure, 0), err
	}

	validator := oryValidation.NewAPIRuleValidator(r.apiRule, r.config.DefaultDomainName)

	return validator.Validate(ctx, client, vsList, gwList), nil
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}
