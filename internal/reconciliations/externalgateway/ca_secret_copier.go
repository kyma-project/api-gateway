package externalgateway

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

// getSecretNamespace returns the namespace for the secret reference
// If the secret reference doesn't specify a namespace, defaults to the ExternalGateway's namespace
func getSecretNamespace(external *externalv1alpha1.ExternalGateway) string {
	if external.Spec.CASecretRef.Namespace != "" {
		return external.Spec.CASecretRef.Namespace
	}
	return external.Namespace
}

// getCACertFromSecret extracts CA certificate data from Secret
// If Secret has exactly one key, uses that key automatically
// If Secret has multiple keys, looks for the expected 'ca.crt' key
func getCACertFromSecret(secret *corev1.Secret, sourceNamespace, sourceName string) ([]byte, error) {
	if len(secret.Data) == 0 {
		return nil, fmt.Errorf("source CA secret %s/%s is empty", sourceNamespace, sourceName)
	}

	if len(secret.Data) == 1 {
		for _, value := range secret.Data {
			ctrl.Log.Info("Using the only available key from CA Secret")
			return value, nil
		}
	}

	cacertData, exists := secret.Data["ca.crt"]
	if !exists {
		return nil, fmt.Errorf("source CA secret %s/%s does not contain 'ca.crt' key (Istio convention)", sourceNamespace, sourceName)
	}
	return cacertData, nil
}

// ReconcileCASecret copies the CA secret from application namespace to istio-system
// Follows Istio convention: uses 'cacert' key for mTLS client certificate validation
func ReconcileCASecret(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway) error {
	sourceNamespace := getSecretNamespace(external)
	sourceName := external.Spec.CASecretRef.Name

	ctrl.Log.Info("Reconciling CA Secret copy", "sourceNamespace", sourceNamespace, "sourceSecret", sourceName)

	// Read source CA secret from the specified namespace
	sourceSecret := &corev1.Secret{}
	sourceKey := types.NamespacedName{
		Name:      sourceName,
		Namespace: sourceNamespace,
	}
	if err := k8sClient.Get(ctx, sourceKey, sourceSecret); err != nil {
		return fmt.Errorf("failed to get source CA secret %s/%s: %w", sourceNamespace, sourceName, err)
	}

	// Extract CA certificate data from source secret
	cacertData, err := getCACertFromSecret(sourceSecret, sourceNamespace, sourceName)
	if err != nil {
		return err
	}

	// Target secret name follows Istio naming convention: <gateway-name>-tls-cacert
	targetSecretName := fmt.Sprintf("%s-tls-cacert", external.GatewayName())

	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetSecretName,
			Namespace: istioSystemNamespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, k8sClient, targetSecret, func() error {
		// Set labels
		targetSecret.Labels = GetStandardLabels(external)

		// Set secret type and data
		targetSecret.Type = corev1.SecretTypeOpaque
		targetSecret.Data = map[string][]byte{
			"ca.crt": cacertData,
		}

		ctrl.Log.Info("Configured CA secret copy", "name", targetSecretName, "namespace", istioSystemNamespace)
		return nil
	})

	return err
}

// DeleteCASecret deletes the CA secret from istio-system
func DeleteCASecret(ctx context.Context, k8sClient client.Client, gatewayName string) error {
	targetSecretName := fmt.Sprintf("%s-tls-cacert", gatewayName)

	ctrl.Log.Info("Deleting CA secret if it exists", "name", targetSecretName, "namespace", istioSystemNamespace)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetSecretName,
			Namespace: istioSystemNamespace,
		},
	}

	err := k8sClient.Delete(ctx, secret)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete CA secret %s/%s: %w", istioSystemNamespace, targetSecretName, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of CA secret as it wasn't present", "name", targetSecretName)
	} else {
		ctrl.Log.Info("Successfully deleted CA secret", "name", targetSecretName)
	}

	return nil
}
