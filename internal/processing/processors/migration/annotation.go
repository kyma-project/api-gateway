package migration

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AnnotationName = "gateway.kyma-project.io/migration-step"

	applyIstioAuthorizationMigrationStep Step = "apply-istio-authorization"
	switchVsToService                    Step = "vs-switch-to-service"
	removeOryRule                        Step = "remove-ory-rule"
)

func ApplyMigrationAnnotation(logger logr.Logger, apiRule client.Object) {
	annotation := nextMigrationStep(apiRule)
	if annotation == removeOryRule {
		logger.Info("Removing migration annotation")
		delete(apiRule.GetAnnotations(), AnnotationName)
	} else {
		logger.Info("Updating migration annotation", "annotation", annotation)
		apiRule.GetAnnotations()[AnnotationName] = string(annotation)
	}
}
