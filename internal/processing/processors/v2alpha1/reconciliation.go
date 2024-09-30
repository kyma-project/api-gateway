package v2alpha1

import (
	"context"
	"errors"

	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/requestauthentication"
	v2alpha1VirtualService "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciliation holds the components needed to reconcile an APIRule. The v2alpha1 reconciliation requires the APIRule in v2alpha1 and v1beta1 since
// not all underlying implementations have been migrated to v2alpha1 and the v1beta1 APIRule is used for those cases.
type Reconciliation struct {
	apiRuleV1beta1  *gatewayv1beta1.APIRule
	apiRuleV2alpha1 *gatewayv2alpha1.APIRule
	processors      []processing.ReconciliationProcessor
	validator       validation.ApiRuleValidator
	config          processing.ReconciliationConfig
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

	if r.validator == nil {
		return make([]validation.Failure, 0), errors.New("validator is not set")
	}
	failures := r.validator.Validate(ctx, client, vsList, gwList)
	return failures, nil
}

func (r Reconciliation) GetProcessors() []processing.ReconciliationProcessor {
	return r.processors
}

func NewReconciliation(apiRuleV2alpha1 *gatewayv2alpha1.APIRule, apiRuleV1beta1 *gatewayv1beta1.APIRule, gateway *networkingv1beta1.Gateway, validator validation.ApiRuleValidator, config processing.ReconciliationConfig, log *logr.Logger, needsMigration bool) Reconciliation {
	var processors []processing.ReconciliationProcessor
	if needsMigration {
		log.Info("APIRule needs migration")
		processors = append(processors, migration.NewMigrationProcessors(apiRuleV2alpha1, apiRuleV1beta1, config, log)...)
	} else {
		processors = append(processors, v2alpha1VirtualService.NewVirtualServiceProcessor(config, apiRuleV2alpha1, gateway))
		processors = append(processors, authorizationpolicy.NewProcessor(log, apiRuleV2alpha1))
		processors = append(processors, requestauthentication.NewProcessor(apiRuleV2alpha1))
	}

	return Reconciliation{
		apiRuleV1beta1:  apiRuleV1beta1,
		apiRuleV2alpha1: apiRuleV2alpha1,
		processors:      processors,
		validator:       validator,
		config:          config,
	}
}
