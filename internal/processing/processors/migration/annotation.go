package migration

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	annotationName = "gateway.kyma-project.io/migration-step"

	applyIstioAuthorizationMigrationStep MigrationStep = "apply-istio-authorization"
	switchVsToService                    MigrationStep = "vs-switch-to-service"
	removeOryRule                        MigrationStep = "remove-ory-rule"
)

func ApplyMigrationAnnotation(ctx context.Context, k8sClient client.Client, apiRule *gatewayv1beta1.APIRule) error {
	annotation := nextMigrationStep(apiRule)
	if annotation == removeOryRule {
		delete(apiRule.Annotations, annotationName)
	} else {
		apiRule.Annotations[annotationName] = string(annotation)
	}
	err := k8sClient.Update(ctx, apiRule)
	if err != nil {
		return err
	}
	return nil
}
