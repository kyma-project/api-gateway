package processing

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing/processor"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IstioReconciliation struct {
	processors []ReconciliationProcessor
	config     ReconciliationConfig
}

func NewIstioReconciliation(config ReconciliationConfig) IstioReconciliation {
	return IstioReconciliation{
		// TODO: Add missing processors for AuthorizationPolicy and RequestAuthentication
		processors: []ReconciliationProcessor{processor.NewIstioVirtualService(config)},
	}
}

func (r IstioReconciliation) validate(apiRule *gatewayv1beta1.APIRule) ([]validation.Failure, error) {

	var vsList networkingv1beta1.VirtualServiceList
	if err := r.config.Client.List(r.config.Ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	validator := validation.APIRule{
		JwtValidator:      &validation.IstioJwtValidator{},
		ServiceBlockList:  r.config.ServiceBlockList,
		DomainAllowList:   r.config.DomainAllowList,
		HostBlockList:     r.config.HostBlockList,
		DefaultDomainName: r.config.DefaultDomainName,
	}
	return validator.Validate(apiRule, vsList), nil
}

func (r IstioReconciliation) getLogger() logr.Logger {
	return r.config.Logger
}

func (r IstioReconciliation) getContext() context.Context {
	return r.config.Ctx
}

func (r IstioReconciliation) getClient() client.Client {
	return r.config.Client
}

func (r IstioReconciliation) getProcessors() []ReconciliationProcessor {
	return r.processors
}
