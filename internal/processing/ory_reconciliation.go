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

type OryReconciliation struct {
	processors []ReconciliationProcessor
	config     ReconciliationConfig
}

func NewOryReconciliation(config ReconciliationConfig) OryReconciliation {
	return OryReconciliation{
		processors: []ReconciliationProcessor{processor.NewOryVirtualService(config), processor.NewAccessRule(config)},
	}
}

func (r OryReconciliation) validate(apiRule *gatewayv1beta1.APIRule) ([]validation.Failure, error) {
	var vsList networkingv1beta1.VirtualServiceList
	if err := r.config.Client.List(r.config.Ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	validator := validation.APIRule{
		JwtValidator:      &validation.OryJwtValidator{},
		ServiceBlockList:  r.config.ServiceBlockList,
		DomainAllowList:   r.config.DomainAllowList,
		HostBlockList:     r.config.HostBlockList,
		DefaultDomainName: r.config.DefaultDomainName,
	}
	return validator.Validate(apiRule, vsList), nil
}

func (r OryReconciliation) getLogger() logr.Logger {
	return r.config.Logger
}

func (r OryReconciliation) getContext() context.Context {
	return r.config.Ctx
}

func (r OryReconciliation) getClient() client.Client {
	return r.config.Client
}

func (r OryReconciliation) getProcessors() []ReconciliationProcessor {
	return r.processors
}
