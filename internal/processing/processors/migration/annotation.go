package migration

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type annotationProcessor struct {
	config processing.ReconciliationConfig
	api    *gatewayv2alpha1.APIRule
}

const (
	annotationName = "gateway.kyma-project.io/migration-step"

	applyIstioAuthorizationMigrationStep MigrationStep = "apply-istio-authorization"
	switchVsToService                    MigrationStep = "vs-switch-to-service"
	removeOryRule                        MigrationStep = "remove-ory-rule"
	finished                             MigrationStep = "migration-finished"
)

func (a annotationProcessor) EvaluateReconciliation(context.Context, client.Client) ([]*processing.ObjectChange, error) {
	annotation := nextMigrationStep(a.api)
	if annotation == removeOryRule {
		delete(a.api.Annotations, annotationName)
	} else {
		a.api.Annotations[annotationName] = string(annotation)
	}

	return []*processing.ObjectChange{processing.NewObjectUpdateAction(a.api)}, nil
}

// NewAnnotationProcessor returns a new instance of the AnnotationProcessor.
func NewAnnotationProcessor(config processing.ReconciliationConfig, api *gatewayv2alpha1.APIRule) processing.ReconciliationProcessor {
	return annotationProcessor{
		api:    api,
		config: config,
	}
}
