package migration

import (
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
)

// NewMigrationProcessors returns a list of processors that should be executed during the migration process.
// Which processors are returned depends on the current migration step indicated by the "api-gateway.kyma-project.io/migration-step" APIRule annotation.
func NewMigrationProcessors(apiRuleV2alpha1 *gatewayv2alpha1.APIRule, apiRuleV1beta1 *gatewayv1beta1.APIRule, config processing.ReconciliationConfig, log *logr.Logger) []processing.ReconciliationProcessor {
	step := nextMigrationStep(apiRuleV1beta1)
	var processors []processing.ReconciliationProcessor
	switch step {
	case removeOryRule: // Step 3
		processors = append(processors, NewAccessRuleDeletionProcessor(config, apiRuleV1beta1))
		fallthrough // We want to also use the processors from the previous steps
	case switchVsToService: // Step 2
		processors = append(processors, virtualservice.NewVirtualServiceProcessor(config, apiRuleV2alpha1))
		fallthrough // We want to also use the processors from the previous steps
	case applyIstioAuthorizationMigrationStep: // Step 1
		processors = append(processors, istio.Newv1beta1AuthorizationPolicyProcessor(config, log, apiRuleV1beta1))
		processors = append(processors, istio.Newv1beta1RequestAuthenticationProcessor(config, apiRuleV1beta1))
	}
	return processors
}

type MigrationStep string

func nextMigrationStep(rule *gatewayv1beta1.APIRule) MigrationStep {
	annotation, found := rule.Annotations[annotationName]
	if !found {
		return applyIstioAuthorizationMigrationStep
	}

	switch annotation {
	case string(applyIstioAuthorizationMigrationStep):
		return switchVsToService
	case string(switchVsToService):
		return removeOryRule
	default:
		return applyIstioAuthorizationMigrationStep
	}
}
