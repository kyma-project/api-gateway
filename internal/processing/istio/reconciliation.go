package istio

import (
	"context"
	"github.com/go-logr/logr"
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

func NewIstioReconciliation(config processing.ReconciliationConfig) Reconciliation {
	vsProcessor := newVirtualService(config)

	return Reconciliation{
		// TODO: Add missing processors for AuthorizationPolicy and RequestAuthentication
		processors: []processing.ReconciliationProcessor{vsProcessor},
		config:     config,
	}
}

func newVirtualService(config processing.ReconciliationConfig) processing.VirtualService {
	return processing.VirtualService{
		Client: config.Client,
		Ctx:    config.Ctx,
		Creator: virtualServiceCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

func (r Reconciliation) Validate(apiRule *gatewayv1beta1.APIRule) ([]validation.Failure, error) {

	var vsList networkingv1beta1.VirtualServiceList
	if err := r.config.Client.List(r.config.Ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	validator := validation.APIRule{
		JwtValidator:      &jwtValidator{},
		ServiceBlockList:  r.config.ServiceBlockList,
		DomainAllowList:   r.config.DomainAllowList,
		HostBlockList:     r.config.HostBlockList,
		DefaultDomainName: r.config.DefaultDomainName,
	}
	return validator.Validate(apiRule, vsList), nil
}

func (r Reconciliation) GetLogger() logr.Logger {
	return r.config.Logger
}

func (r Reconciliation) GetContext() context.Context {
	return r.config.Ctx
}

func (r Reconciliation) GetClient() client.Client {
	return r.config.Client
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}
