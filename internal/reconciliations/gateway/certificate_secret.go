package gateway

import (
	"context"
	_ "embed"
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/version"
)

//go:embed certificate_secret.yaml
var nonGardenerCertificateSecretManifest []byte

func reconcileNonGardenerCertificateSecret(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Certificate Secret", "KymaGatewayEnabled", isEnabled, "name", kymaGatewayCertSecretName, "namespace", certificateDefaultNamespace)

	if !isEnabled || apiGatewayCR.IsInDeletion() {
		return deleteSecret(ctx, k8sClient, kymaGatewayCertSecretName, certificateDefaultNamespace)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = kymaGatewayCertSecretName
	templateValues["Namespace"] = certificateDefaultNamespace
	templateValues["Version"] = version.GetModuleVersion()

	return reconciliations.ApplyResource(ctx, k8sClient, nonGardenerCertificateSecretManifest, templateValues)
}

func deleteSecret(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting certificate secret if it exists", "name", name, "namespace", namespace)
	s := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &s)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete certificate secret %s/%s: %w", certificateDefaultNamespace, name, err)
	}

	ctrl.Log.Info("Successfully deleted certificate secret", "name", name, "namespace", namespace)

	return nil
}
