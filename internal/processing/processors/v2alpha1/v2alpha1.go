package v2alpha1

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	"github.com/kyma-project/api-gateway/internal/validation"
	v2alpha1Validation "github.com/kyma-project/api-gateway/internal/validation/v2alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciliation struct {
	processors []processing.ReconciliationProcessor
	config     processing.ReconciliationConfig
}

func (r Reconciliation) Validate(ctx context.Context, client client.Client, apiRule *gatewayv1beta1.APIRule) ([]validation.Failure, error) {

	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	validator := v2alpha1Validation.APIRuleValidator{
		//HandlerValidator:   &handlerValidator{},
		//InjectionValidator: &injectionValidator{ctx: ctx, client: client},
		DefaultDomainName: r.config.DefaultDomainName,
	}
	return validator.Validate(ctx, client, vsList), nil
}

func (r Reconciliation) GetStatusBase(statusCode string) status.ReconciliationStatusVisitor {
	return Base(statusCode)
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}

func NewReconciliation(config processing.ReconciliationConfig, log *logr.Logger) Reconciliation {
	vsProcessor := istio.NewVirtualServiceProcessor(config)
	apProcessor := istio.NewAuthorizationPolicyProcessor(config, log)
	raProcessor := istio.NewRequestAuthenticationProcessor(config)

	return Reconciliation{
		processors: []processing.ReconciliationProcessor{vsProcessor, raProcessor, apProcessor},
		config:     config,
	}
}

type handlerValidator struct{}

type injectionValidator struct {
	ctx    context.Context
	client client.Client
}
