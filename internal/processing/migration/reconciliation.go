package migration

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/istio"
	"github.com/kyma-project/api-gateway/internal/processing/ory"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciliation struct {
	processors              []processing.ReconciliationProcessor
	config                  processing.ReconciliationConfig
	handlerValidator        validation.HandlerValidator
	accessStrategyValidator validation.AccessStrategyValidator
}

func NewMigrationReconciliation(config processing.ReconciliationConfig, log *logr.Logger) Reconciliation {

	if config.HasMigrationMarker {
		// There is no harm to keep the AccessRules as they are not used. Also, when switching from Oathkeeper VS to Istio VS,
		// there is the risk that there is a short time when the VS is still pointing to Oathkeeper, but the AccessRules are already deleted.
		// Therefore, it's better to keep them for now.
		arProcessor := NewAccessRuleProcessor(config)
		vsProcessor := istio.NewVirtualServiceProcessor(config)
		apProcessor := istio.NewAuthorizationPolicyProcessor(config, log)
		raProcessor := istio.NewRequestAuthenticationProcessor(config)

		return Reconciliation{
			processors:              []processing.ReconciliationProcessor{vsProcessor, raProcessor, apProcessor, arProcessor},
			config:                  config,
			handlerValidator:        handlerValidator{},
			accessStrategyValidator: asValidator{},
		}
	}

	arProcessor := NewAccessRuleProcessor(config)
	vsProcessor := ory.NewVirtualServiceProcessor(config)
	apProcessor := istio.NewAuthorizationPolicyProcessor(config, log)
	raProcessor := istio.NewRequestAuthenticationProcessor(config)

	return Reconciliation{
		processors:              []processing.ReconciliationProcessor{vsProcessor, raProcessor, apProcessor, arProcessor},
		config:                  config,
		handlerValidator:        handlerValidator{},
		accessStrategyValidator: asValidator{},
	}
}

func (r Reconciliation) Validate(ctx context.Context, client client.Client, apiRule *gatewayv1beta1.APIRule) ([]validation.Failure, error) {
	var vsList networkingv1beta1.VirtualServiceList
	if err := client.List(ctx, &vsList); err != nil {
		return make([]validation.Failure, 0), err
	}

	validator := validation.APIRuleValidator{
		HandlerValidator:          r.handlerValidator,
		AccessStrategiesValidator: r.accessStrategyValidator,
		DefaultDomainName:         r.config.DefaultDomainName,
	}
	return validator.Validate(ctx, client, apiRule, vsList), nil
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}

func (r Reconciliation) GetStatusBase(statusCode gatewayv1beta1.StatusCode) processing.ReconciliationStatus {
	if r.config.HasMigrationMarker {
		return istio.StatusBase(statusCode)
	}
	return ory.StatusBase(statusCode)
}

const annotationKey = "gateway.kyma-project.io/migration"
const annotationValue = "v2alpha1"

func (r Reconciliation) ApplyMigrationMarker(apiRule *gatewayv1beta1.APIRule) bool {
	if HasMigrationMarker(*apiRule) {
		return false
	}

	if apiRule.Annotations == nil {
		apiRule.Annotations = map[string]string{}
	}

	apiRule.Annotations[annotationKey] = annotationValue
	return true
}

func HasMigrationMarker(apiRule gatewayv1beta1.APIRule) bool {
	// If the ApiRule is not found, we don't need to do anything. If it's found and converted, CM reconciliation is not needed.
	if apiRule.Annotations != nil {
		if v, ok := apiRule.Annotations[annotationKey]; ok && v == annotationValue {
			ctrl.Log.Info("ApiRule has migration marker to update virtual service.", "name", apiRule.Name, "namespace", apiRule.Namespace)
			return true
		}
	}

	return false
}
