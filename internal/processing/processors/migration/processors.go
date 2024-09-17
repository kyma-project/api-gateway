package migration

import (
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/requestauthentication"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewMigrationProcessors returns a list of processors that should be executed during the migration process.
// Which processors are returned depends on the current migration step indicated by the "api-gateway.kyma-project.io/migration-step" APIRule annotation.
func NewMigrationProcessors(apiRuleV2alpha1 *gatewayv2alpha1.APIRule, apiRuleV1beta1 *gatewayv1beta1.APIRule, config processing.ReconciliationConfig, log *logr.Logger) []processing.ReconciliationProcessor {
	step := nextMigrationStep(apiRuleV1beta1)
	log.Info("Migrating APIRule from v1beta1 to v2alpha1", "step", step)
	var processors []processing.ReconciliationProcessor
	switch step {
	case removeOryRule: // Step 3
		processors = append(processors, NewAccessRuleDeletionProcessor(config, apiRuleV1beta1))
		fallthrough // We want to also use the processors from the previous steps
	case switchVsToService: // Step 2
		processors = append(processors, virtualservice.NewVirtualServiceProcessor(config, apiRuleV2alpha1))
		fallthrough // We want to also use the processors from the previous steps
	case applyIstioAuthorizationMigrationStep: // Step 1
		processors = append(processors, authorizationpolicy.NewMigrationProcessor(log, apiRuleV2alpha1, step != removeOryRule))
		processors = append(processors, requestauthentication.NewProcessor(apiRuleV2alpha1))
	}
	return processors
}

type Step string

func nextMigrationStep(rule client.Object) Step {
	annotations := rule.GetAnnotations()
	annotation, found := annotations[AnnotationName]
	if !found {
		return applyIstioAuthorizationMigrationStep
	}

	switch annotation {
	case string(applyIstioAuthorizationMigrationStep):
		return switchVsToService
	case string(switchVsToService):
		return removeOryRule
	default:
		// applyIstioAuthorizationMigrationStep is used as a fallback in case the annotation is not recognized
		// this ensures that the migration process can continue and the annotation will be corrected if it is invalid
		return applyIstioAuthorizationMigrationStep
	}
}
